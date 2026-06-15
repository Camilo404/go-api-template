# ---- Template API: developer Makefile -----------------------------
# Every target also works with the plain `go` command on Windows if
# `make` is not available. Run `make help` to list targets.

GO        ?= go
BIN_DIR   ?= bin
APP_NAME  ?= api
PKG       := ./...
MAIN_PKG  := ./cmd/api
LDFLAGS   := -trimpath -ldflags="-s -w"

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

SWAG      ?= $(shell $(GO) env GOPATH)/bin/swag

.PHONY: tidy
tidy: ## Sync go.mod / go.sum
	$(GO) mod tidy

.PHONY: swag
swag: ## Generate OpenAPI spec from godoc comments (re-run after changing handler/model comments)
	@if [ ! -x "$(SWAG)" ]; then echo "swag CLI not found. Install: $(GO) install github.com/swaggo/swag/cmd/swag@latest"; exit 1; fi
	$(SWAG) init -g cmd/api/main.go -o docs/ --parseDependency --parseInternal --parseDepth 10

.PHONY: fmt
fmt: ## Format sources
	$(GO) fmt $(PKG)

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKG)

.PHONY: test
test: ## Run unit tests
	$(GO) test -race -count=1 $(PKG)

.PHONY: build
build: ## Build the API binary into ./bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PKG)

.PHONY: run
run: ## Run the API locally
	$(GO) run $(MAIN_PKG)

.PHONY: cover
cover: ## Run tests with coverage report
	$(GO) test -race -coverprofile=coverage.out $(PKG)
	@$(GO) tool cover -func=coverage.out | tail -1

.PHONY: docker-build
docker-build: ## Build the Docker image
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run: ## Run the Docker image locally
	docker run --rm -p 8080:8080 --env-file configs/.env.example $(APP_NAME):latest

.PHONY: clean
clean: ## Remove build artefacts
	rm -rf $(BIN_DIR) coverage.out
