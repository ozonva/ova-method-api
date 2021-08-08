help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 | "sort" }' $(MAKEFILE_LIST)

build: ## Build bin file for current OS
	go build -mod vendor -o ./server ./cmd/ova-method-api/main.go

run: ## Build and run application (go run)
	go run ./cmd/ova-method-api/main.go

test: ## Run tests
	go test -cover -v ./internal
