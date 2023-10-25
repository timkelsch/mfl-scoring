#!/bin/bash

set -euxo pipefail

# Login
aws ecr get-login-password --region us-east-1 | docker login --username AWS \
  --password-stdin 287140326780.dkr.ecr.us-east-1.amazonaws.com

CURRENT_VERSION=$(aws ecr describe-images --region us-east-1 --output json --repository-name mfl-score \
--query 'sort_by(imageDetails,& imagePushedAt)[-1].imageTags[0]' | jq . -r)

IFS=. read -r v1 v2 <<< "${CURRENT_VERSION}"    # split into (integer) components
((v2++))                                        # do the math
NEXT_VERSION="${v1}.${v2}"                      # paste back together

REPO="287140326780.dkr.ecr.us-east-1.amazonaws.com/mfl-score:${NEXT_VERSION}"

IMAGE=$(docker build -q --platform linux/amd64 -t mfl-scoring-image:"${VERSION}" . | cut -d: -f2)
docker tag "${IMAGE}" "${REPO}"
docker push "${REPO}"
