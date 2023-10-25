.PHONY: build

AWS_REGION=us-east-1
AWS_ACCOUNT=$(shell aws sts get-caller-identity | jq -r '.Account')
CODE_DIR=mfl-scoring
IMAGE_URI=$$AWS_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/mfl-score:latest

FUNCTION_NAME=mfl-scoring-check-MflScoringFunction-jsrPurkbiCjK
FUNCTION_VERSION_PROD=31
STACK_NAME=mfl-scoring-check
TEMPLATE_FILE=file://template-check.yaml
 
createstack:
	aws cloudformation create-stack --stack-name ${STACK_NAME} --template-body ${TEMPLATE_FILE} \
		--capabilities CAPABILITY_IAM --region ${AWS_REGION}

updatestack:
	aws cloudformation update-stack --stack-name ${STACK_NAME} --template-body ${TEMPLATE_FILE} \
		--capabilities CAPABILITY_IAM --region ${AWS_REGION}

test:
	cd ${CODE_DIR} && go test -cover

push:
	push.sh

updatelambda:
	aws lambda update-function-code --function-name ${FUNCTION_NAME} \
		--image-uri ${IMAGE_URI}
		--publish --region ${AWS_REGION}

updatestagealias:
	aws lambda update-alias --function-name "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT}:function:${FUNCTION_NAME}" \
		--function-version '$$LATEST' --name STAGE --region ${AWS_REGION}

updateprodalias:
	aws lambda update-alias --function-name "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT}:function:${FUNCTION_NAME}" \
		--function-version ${FUNCTION_VERSION_PROD} --name PROD --region ${AWS_REGION}

pushtostage: test build package push updatelambda updatestagealias

pushtoprod: test build package push updatelambda updateprodalias


val:
	aws cloudformation validate-template --debug --template-body ${TEMPLATE_FILE}

lint:
	cd ${CODE_DIR} && golangci-lint run -v