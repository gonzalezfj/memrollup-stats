.PHONY: all build install clean test lint check-deps

# Variables
BINARY_NAME=memrollup-stats
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
SRC_DIR=src/cmd/memrollup-stats
BIN_DIR=bin

# Check for required tools
check-deps:
	@echo "Checking dependencies..."
	@which go >/dev/null || (echo "Error: Go is not installed. Please install Go first." && exit 1)

all: check-deps build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./$(SRC_DIR)

install: check-deps
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./$(SRC_DIR)

clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	go clean

test: check-deps
	@echo "Running tests..."
	go test -v ./$(SRC_DIR)/...

# Development helpers
dev: build
	@echo "Running in development mode..."
	$(BIN_DIR)/$(BINARY_NAME) -v -F 1 ./test.py

# Release helpers
release: clean build
	@echo "Creating release..."
	tar czf $(BINARY_NAME)-$(VERSION).tar.gz $(BIN_DIR)/$(BINARY_NAME) README.md LICENSE

help:
	@echo "Available targets:"
	@echo "  all        - Build the binary (default)"
	@echo "  build      - Build the binary"
	@echo "  install    - Install the binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  dev        - Build and run with test.py"
	@echo "  release    - Create release archive"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Dependencies:"
	@echo "  - Go (https://golang.org/doc/install)"
	@echo "  - golangci-lint (will be installed automatically)" 