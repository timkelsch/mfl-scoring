.PHONY: build

AWS_REGION=us-east-1
AWS_ACCOUNT=$(shell aws sts get-caller-identity | jq -r '.Account')
BUILD_ARTIFACT=bootstrap.zip
BUILD_DIR=.aws-sam/build
CODE_DIR=mfl-scoring
S3_BUCKET=storage-mflscoring-17z65jydt72tq
S3_PREFIX=builds

API_ID=bc2pcjfiik
FUNCTION_NAME=mfl-scoring-check-MflScoringFunction-jC2WqnR4ihCt
FUNCTION_VERSION_PROD=27
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

build:
	cd ${CODE_DIR} && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o ../.aws-sam/build/bootstrap main.go

package:
	cd ${BUILD_DIR} && zip ${BUILD_ARTIFACT} bootstrap

push:
	cd ${BUILD_DIR} && aws s3 cp ${BUILD_ARTIFACT} s3://${S3_BUCKET}/${S3_PREFIX}/${BUILD_ARTIFACT}

updatelambda:
	aws lambda update-function-code --function-name ${FUNCTION_NAME} \
		--s3-bucket ${S3_BUCKET} \
		--s3-key ${S3_PREFIX}/${BUILD_ARTIFACT} \
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

sam-validate:
	sam validate --lint

lint:
	cd ${CODE_DIR} && golangci-lint run -v

sampush:
	sam build
	sam deploy --no-confirm-changeset

jsambuild:
	/var/jenkins_home/sam/venv/bin/sam build

jsamdeploy:
	/var/jenkins_home/sam/venv/bin/sam deploy -t .aws-sam/build/template.yaml --config-file \
	/var/jenkins_home/workspace/mfl-scoring/samconfig.toml --no-confirm-changeset --no-fail-on-empty-changeset