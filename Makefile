# Nome do executável
BINARY=bin/main

# Diretório de origem principal
SRC_DIR=cmd

# Comandos Go
GO=go
MOCKERY=mockery

# Targets

.PHONY: all
all: build

.PHONY: build
build: generate-code
	$(GO) build -o $(BINARY) ./$(SRC_DIR)/

.PHONY: clean
clean:
	$(GO) clean
	rm -f $(BINARY)
	rm -rf internal/database/mocks
	rm -rf internal/analyzer/oracle/mocks

.PHONY: generate-code
generate-code:
	$(GO) generate ./...

.PHONY: test
test:
	$(GO) test ./...

.PHONY: run
run: build
	./$(BINARY)

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: install-deps
install-deps:
	$(GO) mod download
	$(GO) get github.com/neo4j/neo4j-go-driver/v5
	$(GO) install github.com/vektra/mockery/v2@latest
