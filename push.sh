#!/bin/bash

set -euxo pipefail

# Login
aws ecr get-login-password --region us-east-1 | docker login --username AWS \
  --password-stdin "${AWS_ACCOUNT}.dkr.ecr.us-east-1.amazonaws.com"

# Only works if there is one tag per image
CURRENT_VERSION=$(aws ecr describe-images --region "${AWS_REGION}" --output json --repository-name mfl-score \
--query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags[0]' | jq . -r)

# TODO: Check if the image has changed so we're not just putting a new tag on an existing image

IFS=. read -r v1 v2 <<< "${CURRENT_VERSION}"    # split into (integer) components
((v2++))                                        # do the math
NEXT_VERSION="${v1}.${v2}"                      # paste back together
# NEXT_VERSION=0.13

NEW_IMAGE_URI="${AWS_ACCOUNT}.dkr.ecr.us-east-1.amazonaws.com/mfl-score:${NEXT_VERSION}"

NEW_IMAGE_SHA=$(docker build -q -t mfl-scoring-image:"${NEXT_VERSION}" . | cut -d: -f2)

CURRENT_VERSION_SHA=$(aws ecr describe-images --repository-name mfl-score --image-ids imageTag="${CURRENT_VERSION}" | \
  jq -r '.imageDetails[] | select(.imageTags[]=="'"${CURRENT_VERSION}"'") | .imageDigest' | cut -d: -f2)

if [[ "${CURRENT_VERSION_SHA}" == "${NEW_IMAGE_SHA}" ]]; then
  exit
fi

docker tag "${NEW_IMAGE_SHA}" "${NEW_IMAGE_URI}"
docker push "${NEW_IMAGE_URI}"

aws lambda update-function-code --function-name "${FUNCTION_NAME}" --architectures arm64 \
		--image-uri "${NEW_IMAGE_URI}" --publish --region "${AWS_REGION}"