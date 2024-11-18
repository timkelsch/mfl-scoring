#!/bin/bash

set -euxo pipefail

echo "${FUNCTION_NAME}"

# Get lambda version of STAGE alias
STAGE_VERSION=$(aws lambda get-alias --function-name "${FUNCTION_NAME}" --name STAGE | jq -r '.FunctionVersion')

# STAGE_VERSION == $LATEST, then check which lambda version is the most recent
if [ "${STAGE_VERSION}" == '$LATEST' ]; then
    # Get lambda version of PROD alias  
    STAGE_VERSION=$(aws lambda list-versions-by-function --function-name "${FUNCTION_NAME}" \
      --query "max_by(Versions, &to_number(to_number(Version) || '0'))" | jq -r '.Version')
fi

# set PROD alias to STAGE version
aws lambda update-alias --function-name "${FUNCTION_NAME}" --name PROD --function-version "${STAGE_VERSION}" \
  || echo "ERROR: Could not update alias PROD to version ${STAGE_VERSION}"