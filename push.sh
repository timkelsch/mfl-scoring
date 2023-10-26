#!/bin/bash

set -euxo pipefail

# Login
aws ecr get-login-password --region us-east-1 | docker login --username AWS \
  --password-stdin "${AWS_ACCOUNT}.dkr.ecr.us-east-1.amazonaws.com"

CURRENT_VERSION=$(aws ecr describe-images --region "${AWS_REGION}" --output json --repository-name mfl-score \
--query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags[0]' | jq . -r)

# TODO: Check if the image has changed so we're not just putting a new tag on an existing image

IFS=. read -r v1 v2 <<< "${CURRENT_VERSION}"    # split into (integer) components
((v2++))                                        # do the math
NEXT_VERSION="${v1}.${v2}"                      # paste back together

IMAGE_URI="${AWS_ACCOUNT}.dkr.ecr.us-east-1.amazonaws.com/mfl-score:${NEXT_VERSION}"

IMAGE=$(docker build -q -t mfl-scoring-image:"${NEXT_VERSION}" . | cut -d: -f2)
docker tag "${IMAGE}" "${IMAGE_URI}"
docker push "${IMAGE_URI}"

aws lambda update-function-code --function-name "${FUNCTION_NAME}" \
		--image-uri "${IMAGE_URI}" --publish --region "${AWS_REGION}"