#!/bin/bash

VERSION=$1
REPO="287140326780.dkr.ecr.us-east-1.amazonaws.com/mfl-score:${VERSION}"

IMAGE=$(docker build -q --platform linux/amd64 -t mfl-scoring-image:0.7 . | cut -d: -f2)
docker tag ${IMAGE} ${REPO}
docker push ${REPO}
