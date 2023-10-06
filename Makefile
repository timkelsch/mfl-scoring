.PHONY: build

build:
	cd mfl-scoring
	go build

test:
	cd mfl-scoring
	go test -v -cover

validate:
	sam validate --lint

lint:
	cd mfl-scoring && golangci-lint run -v

sampush:
	sam build
	sam deploy --no-confirm-changeset

jsambuild:
	/var/jenkins_home/sam/venv/bin/sam build

jsamdeploy:
	/var/jenkins_home/sam/venv/bin/sam deploy -t .aws-sam/build/template.yaml --config-file \
	/var/jenkins_home/workspace/mfl-pipeline/samconfig.toml --no-confirm-changeset --no-fail-on-empty-changeset