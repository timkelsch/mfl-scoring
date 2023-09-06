.PHONY: build

build:
	sam build

validate:
	sam validate --lint

sampush:
	sam build
	sam deploy --no-confirm-changeset