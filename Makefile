# ============================================
# Duct - Pipeline as Code
# ============================================

BINARY_NAME := duct
BUILD_DIR := ./build
INSTALL_DIR := /usr/local/bin
GO := go

# ============================================
# COLORS
# ============================================

GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
RESET := \033[0m

# ============================================
# DEFAULT
# ============================================

.DEFAULT_GOAL := help

# ============================================
# TARGETS
# ============================================

.PHONY: help
help: ## Show this help
	@printf '%b\n' "$(GREEN)Duct - Pipeline as Code$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the duct binary
	@printf '%b\n' "$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X main.version=dev -X main.commit=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/duct
	@printf '%b\n' "$(GREEN)Built: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: install
install: build ## Install duct to /usr/local/bin
	@printf '%b\n' "$(GREEN)Installing to $(INSTALL_DIR)...$(RESET)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@printf '%b\n' "$(GREEN)Installed! Run: duct --version$(RESET)"

.PHONY: uninstall
uninstall: ## Remove duct from /usr/local/bin
	@printf '%b\n' "$(YELLOW)Removing from $(INSTALL_DIR)...$(RESET)"
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@printf '%b\n' "$(GREEN)Uninstalled!$(RESET)"

.PHONY: run
run: ## Run duct (dev mode)
	$(GO) run ./cmd/duct $(ARGS)

.PHONY: test
test: ## Run all tests
	@printf '%b\n' "$(GREEN)Running tests...$(RESET)"
	$(GO) test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@printf '%b\n' "$(GREEN)Running tests with coverage...$(RESET)"
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@printf '%b\n' "$(GREEN)Coverage report: coverage.html$(RESET)"

.PHONY: clean
clean: ## Clean build artifacts
	@printf '%b\n' "$(YELLOW)Cleaning...$(RESET)"
	@rm -rf $(BUILD_DIR) coverage.out coverage.html
	@printf '%b\n' "$(GREEN)Cleaned!$(RESET)"

.PHONY: fmt
fmt: ## Format Go code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: fmt vet ## Run all linters

.PHONY: deps
deps: ## Download and verify dependencies
	$(GO) mod download
	$(GO) mod verify

.PHONY: tidy
tidy: ## Tidy go modules
	$(GO) mod tidy

.PHONY: release
release: clean ## Build for multiple platforms
	@printf '%b\n' "$(GREEN)Building releases...$(RESET)"
	@mkdir -p $(BUILD_DIR)/release
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "-s -w" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 ./cmd/duct
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "-s -w" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 ./cmd/duct
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "-s -w" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 ./cmd/duct
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags "-s -w" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 ./cmd/duct
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "-s -w" -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe ./cmd/duct
	@printf '%b\n' "$(GREEN)Release binaries in $(BUILD_DIR)/release/$(RESET)"

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t duct:latest .

.PHONY: local-test
local-test: build ## Test duct locally with example Ductfile
	@printf '%b\n' "$(GREEN)Running local test...$(RESET)"
	$(BUILD_DIR)/$(BINARY_NAME) run --local --file Ductfile.example