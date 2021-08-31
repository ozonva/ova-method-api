help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 | "sort" }' $(MAKEFILE_LIST)

build: ## Build bin file for current OS
	@go build -mod vendor -o ./server ./cmd/ova-method-api/main.go

run: ## Build and run application (go run)
	@go run ./cmd/ova-method-api/main.go

gen: gen-proto ## Code generation
	@go generate ./...

gen-proto:
	@protoc \
	--go_out=./pkg/ova-method-api --go_opt=paths=import \
	--go-grpc_out=./pkg/ova-method-api --go-grpc_opt=paths=import \
	./api/ova-method-api/service.proto

test: ## Run tests
	@go test -cover -race -v ./...

deps: ## Install service dependencies
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
