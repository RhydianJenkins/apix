help: ## Show this help
	@grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-50s\033[0m %s\n", $$1, $$2}'

build: ## Build apix into a binary
	@nix run .#build

test: ## Run all tests
	@nix run .#test

vet: ## Run go vet
	@nix run .#vet

release: build vet test ## Tag and push a release using the version in VERSION
	@version=v$(shell cat VERSION); \
	if git rev-parse "$$version" >/dev/null 2>&1; then \
		echo "Tag $$version already exists"; \
		exit 1; \
	fi; \
	git tag -a "$$version" -m "Release $$version"; \
	git push origin "$$version"; \
	echo "Released $$version"
