BINARY      = bin/deptree
SRC_DIR     = ./cmd/deptree
GO          = go
CGO_ENABLED = 0

# System name used when importing (override with: make docker-import SYSTEM=myapp)
SYSTEM ?= petclinic

DOCKER         = cd docker && docker compose
DOCKER_ORACLE  = cd docker && docker compose --profile oracle
NEO4J_CLEAR    = docker exec deptree-neo4j cypher-shell -u neo4j -p deptree123 "MATCH (n) DETACH DELETE n"

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: all
all: build ## Default: build

.PHONY: build
build: ## Compile the binary to bin/deptree
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BINARY) $(SRC_DIR)

.PHONY: install
install: ## Install deptree to $GOPATH/bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install $(SRC_DIR)

.PHONY: run
run: build ## Build and run with --help
	./$(BINARY) --help

.PHONY: clean
clean: ## Remove build artifacts
	$(GO) clean
	rm -f $(BINARY)

# ── Quality ───────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests (CGO_ENABLED=0 avoids dyld issues on macOS)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -v ./...

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

.PHONY: install-deps
install-deps: ## Install development tools (golangci-lint, mockery)
	$(GO) mod download
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	$(GO) install github.com/vektra/mockery/v2@latest

# ── Docker — environment ──────────────────────────────────────────────────────

.PHONY: docker-up
docker-up: ## Start Neo4j (and clone spring-petclinic source)
	$(DOCKER) up -d neo4j petclinic-init

.PHONY: docker-up-oracle
docker-up-oracle: ## Start Neo4j + Oracle Free (first boot takes ~3 min)
	$(DOCKER_ORACLE) up -d neo4j oracle

.PHONY: docker-down
docker-down: ## Stop all containers
	$(DOCKER_ORACLE) down

.PHONY: docker-down-clean
docker-down-clean: ## Stop all containers and wipe volumes (full reset)
	$(DOCKER_ORACLE) down -v

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild deptree image without cache (run after code changes)
	$(DOCKER_ORACLE) build --no-cache deptree deptree-oracle

# ── Docker — data ─────────────────────────────────────────────────────────────

.PHONY: docker-neo4j-clear
docker-neo4j-clear: ## Delete all nodes and edges from Neo4j (containers stay up)
	$(NEO4J_CLEAR)

.PHONY: docker-oracle-schema
docker-oracle-schema: ## Create/recreate petclinic Oracle schema (tables, views, procedures, package)
	$(DOCKER_ORACLE) up oracle-petclinic-init

# ── Docker — import ───────────────────────────────────────────────────────────

.PHONY: docker-import-java
docker-import-java: ## Import spring-petclinic Java source into Neo4j
	$(DOCKER) run --rm deptree \
		--system=$(SYSTEM) \
		files java \
		--root-dir=/workspace/spring-petclinic/src/main/java \
		--db-schema=system

.PHONY: docker-import-oracle
docker-import-oracle: ## Set up Oracle petclinic schema and import into Neo4j
	$(DOCKER_ORACLE) up oracle-petclinic-init
	$(DOCKER_ORACLE) run --rm deptree-oracle

.PHONY: docker-import
docker-import: ## Import Java + Oracle (full petclinic graph) — requires Oracle running
	$(DOCKER_ORACLE) run --rm deptree --system=$(SYSTEM) all

.PHONY: docker-reimport
docker-reimport: ## Clear Neo4j and reimport everything from scratch
	$(NEO4J_CLEAR)
	$(MAKE) docker-import

# ── Docker — web UI ───────────────────────────────────────────────────────────

.PHONY: docker-serve
docker-serve: ## Start the web UI at http://localhost:8090
	$(DOCKER) up deptree-serve
