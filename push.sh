#!/bin/bash

set -euxo pipefail

REGISTRY="${AWS_ACCOUNT}.dkr.ecr.us-east-1.amazonaws.com"
REPO="mfl-score"

# Login - not necessary with credential helper
# aws ecr get-login-password --region us-east-1 | docker login --username AWS \
#   --password-stdin "${REGISTRY}"

# Only works if there is one tag per image
CURRENT_VERSION=$(aws ecr describe-images --region "${AWS_REGION}" --output json --repository-name mfl-score \
--query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags[0]' | jq . -r)
CURRENT_IMAGE="${REGISTRY}/${REPO}:${CURRENT_VERSION}"

IFS=. read -r v1 v2 <<< "${CURRENT_VERSION}"    # split into (integer) components
((v2++))                                        # do the math
NEXT_VERSION="${v1}.${v2}"                      # paste back together
# NEXT_VERSION=0.15

NEW_IMAGE_URI="${REGISTRY}/${REPO}:${NEXT_VERSION}"

NEW_IMAGE_ID=$(docker build -q -t "${REPO}:${NEXT_VERSION}" . | cut -d: -f2 | head -c 12)

# Check if CURRENT_IMAGE already exists locally
if [[ $(docker image ls --format json "${CURRENT_IMAGE}" | jq -r '.ID' | wc -l) -eq 1 ]]; then
  # If so, set CURRENT_IMAGE_ID
  CURRENT_IMAGE_ID=$(docker image ls --format json "${CURRENT_IMAGE}" | jq -r '.ID');
  else
    # If not, pull the current image from the repo
    docker pull "${CURRENT_IMAGE}"
    # And set the CURRENT_IMAGE_ID using that
    if ! CURRENT_IMAGE_ID=$(docker inspect --format '{{.Id}}' "${CURRENT_IMAGE}"); then
      echo "Cannot determine current image ID. Exiting."
      exit 2
    fi
fi

if [ "${CURRENT_IMAGE_ID}" = "${NEW_IMAGE_ID}" ]; then
  echo "The image built for this commit is the same as the most recent image in the remote repository. Exiting."
  exit 3
fi

docker tag "${NEW_IMAGE_ID}" "${NEW_IMAGE_URI}"
docker push "${NEW_IMAGE_URI}"

aws lambda update-function-code --function-name "${FUNCTION_NAME}" --architectures arm64 \
		--image-uri "${NEW_IMAGE_URI}" --publish --region "${AWS_REGION}"

# This gets complicated because if an image has been uploaded before, even if it has since been deleted,
# ECR will use the imagePushedAt of the first time it was pushed. This breaks the sort by query in 
# CURRENT_VERSION.