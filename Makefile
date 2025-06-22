help: ## Show this help
	@grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-50s\033[0m %s\n", $$1, $$2}'

build: ## Build apix into a binary
	@go build -o apix ./main.go

test: ## Run all tests
	@go test -v ./...
