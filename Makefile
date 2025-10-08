# Makefile for go-mcp-atlassian
.PHONY: help build clean test test-coverage lint fmt vet tidy install run dev docker-build docker-run

# Variables
BINARY_NAME=atlas-mcp
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR=bin
MAIN_PATH=./cmd/atlas-mcp
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Platform-specific variables
ifeq ($(OS),Windows_NT)
	BINARY_EXTENSION=.exe
else
	BINARY_EXTENSION=
endif

# Colors for terminal output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

## help: Display this help message
help:
	@echo "$(COLOR_BOLD)go-mcp-atlassian - Available targets:$(COLOR_RESET)"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  $(COLOR_GREEN)/' | sed 's/:/ $(COLOR_RESET)-/'
	@echo ""

## build: Build the application binary
build:
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) $(MAIN_PATH)
	@echo "$(COLOR_GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION)$(COLOR_RESET)"

## build-all: Build for all platforms (Linux, macOS, Windows)
build-all: clean
	@echo "$(COLOR_BLUE)Building for all platforms...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "$(COLOR_GREEN)✓ Built binaries for all platforms in $(BUILD_DIR)/$(COLOR_RESET)"

## install: Install the binary to $GOPATH/bin
install:
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)
	@echo "$(COLOR_GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

## clean: Remove build artifacts and cache
clean:
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)$(BINARY_EXTENSION)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

## test: Run all tests
test:
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	$(GOTEST) -v ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

## test-short: Run tests with short flag (skip long-running tests)
test-short:
	@echo "$(COLOR_BLUE)Running short tests...$(COLOR_RESET)"
	$(GOTEST) -short -v ./...
	@echo "$(COLOR_GREEN)✓ Short tests passed$(COLOR_RESET)"

## test-race: Run tests with race detector
test-race:
	@echo "$(COLOR_BLUE)Running tests with race detector...$(COLOR_RESET)"
	$(GOTEST) -race -v ./...
	@echo "$(COLOR_GREEN)✓ Race tests passed$(COLOR_RESET)"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

## bench: Run benchmarks
bench:
	@echo "$(COLOR_BLUE)Running benchmarks...$(COLOR_RESET)"
	$(GOTEST) -bench=. -benchmem ./...

## lint: Run golangci-lint
lint:
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo "$(COLOR_GREEN)✓ Linting passed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

## fmt: Format Go code
fmt:
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	$(GOFMT) ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

## vet: Run go vet
vet:
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	$(GOVET) ./...
	@echo "$(COLOR_GREEN)✓ Vet passed$(COLOR_RESET)"

## tidy: Tidy and verify go.mod
tidy:
	@echo "$(COLOR_BLUE)Tidying modules...$(COLOR_RESET)"
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "$(COLOR_GREEN)✓ Modules tidied$(COLOR_RESET)"

## deps: Download dependencies
deps:
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	$(GOMOD) download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

## update-deps: Update all dependencies
update-deps:
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

## run: Build and run the application
run: build
	@echo "$(COLOR_BLUE)Running $(BINARY_NAME)...$(COLOR_RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION)

## dev: Run the application in development mode with verbose logging
dev:
	@echo "$(COLOR_BLUE)Running in development mode...$(COLOR_RESET)"
	MCP_VERY_VERBOSE=true $(GOCMD) run $(MAIN_PATH)

## check: Run all quality checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

## docker-build: Build Docker image
docker-build:
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "$(COLOR_GREEN)✓ Docker image built$(COLOR_RESET)"

## docker-run: Run Docker container
docker-run:
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	docker run --rm -it --env-file .env $(BINARY_NAME):latest

## version: Show version information
version:
	@echo "$(COLOR_BOLD)Version:$(COLOR_RESET) $(VERSION)"
	@echo "$(COLOR_BOLD)Go version:$(COLOR_RESET) $(shell go version)"
	@echo "$(COLOR_BOLD)Build dir:$(COLOR_RESET) $(BUILD_DIR)"

## size: Show binary size
size: build
	@echo "$(COLOR_BOLD)Binary size:$(COLOR_RESET)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) | awk '{print $$5 "\t" $$9}'

## info: Show project information
info:
	@echo "$(COLOR_BOLD)Project Information:$(COLOR_RESET)"
	@echo "  Binary name: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build directory: $(BUILD_DIR)"
	@echo "  Main package: $(MAIN_PATH)"
	@echo "  Go version: $(shell go version | awk '{print $$3}')"
	@echo ""
	@echo "$(COLOR_BOLD)Project Statistics:$(COLOR_RESET)"
	@echo "  Go files: $(shell find . -name '*.go' -not -path './vendor/*' | wc -l | tr -d ' ')"
	@echo "  Lines of code: $(shell find . -name '*.go' -not -path './vendor/*' -exec wc -l {} + | tail -1 | awk '{print $$1}')"
	@echo "  Packages: $(shell go list ./... | wc -l | tr -d ' ')"

## tools: Install development tools
tools:
	@echo "$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(COLOR_GREEN)✓ Development tools installed$(COLOR_RESET)"

## pre-commit: Run pre-commit checks (fmt, vet, lint, test)
pre-commit: fmt vet lint test-short
	@echo "$(COLOR_GREEN)✓ Pre-commit checks passed$(COLOR_RESET)"

# Default target
.DEFAULT_GOAL := help
