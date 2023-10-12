.PHONY: build

AWS_REGION=us-east-1
BUILD_ARTIFACT=bootstrap.zip
BUILD_DIR=.aws-sam/build
CODE_DIR=mfl-scoring
S3_BUCKET=mfl-scoring-builds
S3_PREFIX=builds
FUNCTION_NAME=
STACK_NAME=mfl-scoring-http
TEMPLATE_FILE=file://template-http.yaml
BUILD_NUMBER=one

createstack:
	aws cloudformation create-stack --stack-name ${STACK_NAME} --template-body ${TEMPLATE_FILE} \
		--capabilities CAPABILITY_IAM --region ${AWS_REGION}

updatestack:
	aws cloudformation update-stack --stack-name ${STACK_NAME} --template-body ${TEMPLATE_FILE} \
		--capabilities CAPABILITY_IAM --region ${AWS_REGION}

test:
	cd ${CODE_DIR} && go test -v -cover

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
	aws lambda update-alias --function-name ${FUNCTION_NAME} --name STAGE --region ${AWS_REGION}

pushtostage: test build package push updatelambda updatestagealias


addpermission:
	aws lambda add-permission --function-name ${FUNCTION_NAME} --statement-id ${BUILD_NUMBER} \
		--action lambda:InvokeFunction --principal apigateway.amazonaws.com

# publishversion:
# 	aws lambda publish-version --function-name ${FUNCTION_NAME} \
# 		--description ${BUILD_NUMBER} --region ${AWS_REGION}

# updatestagealias-stupid:
# 	LAMBDA_VERSION=$(shell aws lambda list-versions-by-function --function-name ${FUNCTION_NAME} \
# 		--region ${AWS_REGION} --output json | jq -r ".Versions[] | select(.Version!=\"\\\$LATEST\") \
# 		| select(.Description == \"${BUILD_NUMBER}\").Version"); \
# 	aws lambda update-alias --function-name ${FUNCTION_NAME} --name STAGE --function-version \
# 		$${LAMBDA_VERSION} --description ${BUILD_NUMBER} --region ${AWS_REGION}

val:
	aws cloudformation validate-template --template-body ${TEMPLATE_FILE}

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