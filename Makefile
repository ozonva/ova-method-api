DB_NAME=ova
DB_USER=root
DB_PASS=123456

help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 | "sort" }' $(MAKEFILE_LIST)

build: ## Build bin file for current OS
	@go build -mod vendor -o ./server ./cmd/ova-method-api/main.go

run: ## Build and run application (go run)
	@go run ./cmd/ova-method-api/main.go

gen: ## Code generation
	@go generate ./...
	@protoc \
	--go_out=./pkg/ova-method-api --go_opt=paths=import \
	--go-grpc_out=./pkg/ova-method-api --go-grpc_opt=paths=import \
	./api/ova-method-api/service.proto

test: ## Run tests
	@go test -cover -race -v ./...

lint: ## Run linter
	@./bin/golangci-lint run ./...

deps: ## Install service dependencies
	@go install github.com/pressly/goose/v3/cmd/goose@v3.1
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.42.0

goose: ## Migration manager. Example: make goose cmd="-h"
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING="user=${DB_USER} password=${DB_PASS} dbname=${DB_NAME} sslmode=disable" goose -table migrations -dir ./migrations $(cmd)

start: ## Run environment
	@DB_NAME=${DB_NAME} DB_USER=${DB_USER} DB_PASS=${DB_PASS} docker-compose -f ./deploy/docker-compose.yaml up -d

stop: ## Stop environment
	@docker-compose -f ./deploy/docker-compose.yaml stop
