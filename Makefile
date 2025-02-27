default:  build

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run unit tests
	@echo "+ $@"
	go test -cover -race -coverprofile=coverage.tmp ./pkg/...

.PHONY: build
build: generate-api test

.PHONY: generate-api
generate-api:
	go generate ./...
