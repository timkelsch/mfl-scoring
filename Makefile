.PHONY: build

build:
	cd mfl-scoring
	go build

validate:
	sam validate --lint

sampush:
	sam build
	sam deploy --no-confirm-changeset