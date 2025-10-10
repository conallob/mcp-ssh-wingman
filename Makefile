.PHONY: build clean install test run

BINARY_NAME=mcp-ssh-wingman
BUILD_DIR=bin
INSTALL_PATH=/usr/local/bin

# Get version from git tag, or use "dev" if not on a tag
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/conall-obrien/mcp-ssh-wingman/internal/server.ServerVersion=$(VERSION)"

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mcp-ssh-wingman

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	@go test -v ./...

run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Vetting code..."
	@go vet ./...

lint: fmt vet
	@echo "Linting complete"
