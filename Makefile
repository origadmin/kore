# Makefile for Kore DI Library

.PHONY: all test lint fmt clean help

all: fmt lint test

# Run all tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format all go files
fmt:
	@echo "Formatting files..."
	go fmt ./...
	go mod tidy

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f coverage.out

# Display available commands
help:
	@echo "Available commands:"
	@echo "  make test  - Run tests with race detection and coverage"
	@echo "  make lint  - Run golangci-lint"
	@echo "  make fmt   - Format code and tidy mod files"
	@echo "  make clean - Remove temporary files"
	@echo "  make all   - Run fmt, lint, and test"
