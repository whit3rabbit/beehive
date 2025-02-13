# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=manager
MAIN_PATH=./cmd/manager

# Build flags
BUILD_FLAGS=-ldflags="-s -w"
RACE_FLAGS=-race

# Docker parameters
DOCKER_IMAGE=beehive-manager
DOCKER_TAG=latest

# Directory structure
BIN_DIR=bin
DIST_DIR=dist
CERT_DIR=certs

# Color output
COLOR_RESET=\033[0m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m

.PHONY: all build clean test coverage lint fmt deps tidy docker docker-compose help setup

all: clean build test ## Default target: clean, build and test

build: ## Build the application
	@echo "${COLOR_GREEN}Building...${COLOR_RESET}"
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)

build-race: ## Build with race detector
	@echo "${COLOR_GREEN}Building with race detector...${COLOR_RESET}"
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(RACE_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build ## Run the application
	@echo "${COLOR_GREEN}Running...${COLOR_RESET}"
	./$(BIN_DIR)/$(BINARY_NAME)

clean: ## Clean build artifacts
	@echo "${COLOR_GREEN}Cleaning...${COLOR_RESET}"
	@rm -rf $(BIN_DIR)
	@rm -rf $(DIST_DIR)
	$(GOCLEAN)

test: ## Run tests
	@echo "${COLOR_GREEN}Running tests...${COLOR_RESET}"
	$(GOTEST) -v ./...

test-race: ## Run tests with race detector
	@echo "${COLOR_GREEN}Running tests with race detector...${COLOR_RESET}"
	$(GOTEST) $(RACE_FLAGS) -v ./...

coverage: ## Generate test coverage report
	@echo "${COLOR_GREEN}Generating coverage report...${COLOR_RESET}"
	@mkdir -p $(DIST_DIR)
	$(GOTEST) -coverprofile=$(DIST_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(DIST_DIR)/coverage.out -o $(DIST_DIR)/coverage.html
	@echo "${COLOR_YELLOW}Coverage report generated at $(DIST_DIR)/coverage.html${COLOR_RESET}"

lint: ## Run linter
	@echo "${COLOR_GREEN}Running linter...${COLOR_RESET}"
	$(GOCMD) vet ./...
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "${COLOR_YELLOW}golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${COLOR_RESET}"; \
	fi

fmt: ## Format code
	@echo "${COLOR_GREEN}Formatting code...${COLOR_RESET}"
	$(GOCMD) fmt ./...

deps: ## Download dependencies
	@echo "${COLOR_GREEN}Downloading dependencies...${COLOR_RESET}"
	$(GOMOD) download

tidy: ## Tidy up module dependencies
	@echo "${COLOR_GREEN}Tidying up modules...${COLOR_RESET}"
	$(GOMOD) tidy

setup: ## Run setup command
	@echo "${COLOR_GREEN}Running setup...${COLOR_RESET}"
	@mkdir -p $(CERT_DIR)
	./$(BIN_DIR)/$(BINARY_NAME) setup

# Cross compilation targets
.PHONY: build-all build-linux build-windows build-darwin

build-all: build-linux build-windows build-darwin ## Build for all platforms

build-linux: ## Build for Linux
	@echo "${COLOR_GREEN}Building for Linux...${COLOR_RESET}"
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-windows: ## Build for Windows
	@echo "${COLOR_GREEN}Building for Windows...${COLOR_RESET}"
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-darwin: ## Build for macOS
	@echo "${COLOR_GREEN}Building for macOS...${COLOR_RESET}"
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

# Docker targets
.PHONY: docker docker-compose

docker: ## Build Docker image
	@echo "${COLOR_GREEN}Building Docker image...${COLOR_RESET}"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-compose: ## Run with Docker Compose
	@echo "${COLOR_GREEN}Starting services with Docker Compose...${COLOR_RESET}"
	docker-compose up --build

# Help target
help: ## Display this help
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${COLOR_GREEN}%-20s${COLOR_RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)