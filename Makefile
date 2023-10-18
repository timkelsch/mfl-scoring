.PHONY: build

AWS_REGION=us-east-1
AWS_ACCOUNT=287140326780
BUILD_ARTIFACT=bootstrap.zip
BUILD_DIR=.aws-sam/build
CODE_DIR=mfl-scoring
S3_BUCKET=mfl-scoring-builds
S3_PREFIX=builds
API_ID=cxotw5q60g
FUNCTION_NAME_STAGE=mfl-scoring-http-MflScoringStageFunction-cgdT1EMOrMbd
FUNCTION_NAME_PROD=mfl-scoring-http-MflScoringProdFunction-vZajLTdguZE7
FUNCTION_VERSION_STAGE=4
FUNCTION_VERSION_PROD=2
STACK_NAME=mfl-scoring-http
TEMPLATE_FILE=file://template-http.yaml

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

updatelambdastage:
	aws lambda update-function-code --function-name ${FUNCTION_NAME_STAGE} \
		--s3-bucket ${S3_BUCKET} \
		--s3-key ${S3_PREFIX}/${BUILD_ARTIFACT} \
		--publish --region ${AWS_REGION}

updatelambdaprod:
	aws lambda update-function-code --function-name ${FUNCTION_NAME_PROD} \
		--s3-bucket ${S3_BUCKET} \
		--s3-key ${S3_PREFIX}/${BUILD_ARTIFACT} \
		--publish --region ${AWS_REGION}

updatestagealias:
	aws lambda update-alias --function-name "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT}:function:${FUNCTION_NAME_STAGE}" \
		--function-version ${FUNCTION_VERSION_STAGE} --name STAGE --region ${AWS_REGION}

updateprodalias:
	aws lambda update-alias --function-name "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT}:function:${FUNCTION_NAME_PROD}" \
		--function-version ${FUNCTION_VERSION_PROD} --name PROD --region ${AWS_REGION}

pushtostage: test build package push updatelambdastage updatestagealias

pushtoprod: test build package push updatelambdaprod updateprodalias


addpermissionstage:
	aws lambda add-permission --function-name "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT}:function:${FUNCTION_NAME_STAGE}:${FUNCTION_VERSION_STAGE}" \
		 --statement-id ${shell uuidgen} --action lambda:InvokeFunction --principal apigateway.amazonaws.com \
		 --source-arn arn:aws:execute-api:${AWS_REGION}:${AWS_ACCOUNT}:${API_ID}/*/GET/*

#		--qualifier STAGE --statement-id ${shell uuidgen} --action lambda:InvokeFunction --principal apigateway.amazonaws.com

addpermissionprod:
	aws lambda add-permission --function-name "arn:aws:lambda:${AWS_REGION}:${AWS_ACCOUNT}:function:${FUNCTION_NAME_PROD}:${FUNCTION_VERSION_PROD}" \
		--statement-id ${shell uuidgen} --action lambda:InvokeFunction --principal apigateway.amazonaws.com
		--source-arn arn:aws:execute-api:${AWS_REGION}:${AWS_ACCOUNT}:${API_ID}/*/GET/*

#		--qualifier PROD --statement-id ${shell uuidgen} --action lambda:InvokeFunction --principal apigateway.amazonaws.com


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