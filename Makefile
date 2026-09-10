help: ## Show this help
	@grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-50s\033[0m %s\n", $$1, $$2}'

build: ## Build apix into a binary
	@go build -ldflags "-X main.version=$(shell cat VERSION)" -o apix ./cmd/apix

test: ## Run all tests
	@go test -v ./...
