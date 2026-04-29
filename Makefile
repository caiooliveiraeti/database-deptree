BINARY     = bin/deptree
SRC_DIR    = ./cmd/deptree
GO         = go
CGO_ENABLED = 0

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build ## Default: build

.PHONY: build
build: ## Compile the binary to bin/deptree
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BINARY) $(SRC_DIR)

.PHONY: install
install: ## Install deptree to $GOPATH/bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install $(SRC_DIR)

.PHONY: test
test: ## Run tests (CGO_ENABLED=0 to avoid dyld issues on macOS)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -v ./...

.PHONY: run
run: build ## Build and run with --help
	./$(BINARY) --help

.PHONY: fmt
fmt: ## Format Go source files
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	CGO_ENABLED=$(CGO_ENABLED) $(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: generate-mocks
generate-mocks: ## Regenerate mockery mocks
	$(GO) generate ./...

.PHONY: clean
clean: ## Remove build artifacts
	$(GO) clean
	rm -f $(BINARY)

.PHONY: install-deps
install-deps: ## Install development tools
	$(GO) mod download
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	$(GO) install github.com/vektra/mockery/v2@latest
