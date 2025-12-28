.PHONY: build run clean test help

# Build the server
build:
	@echo "Building pertisk-chart..."
	@go build -ldflags="-s -w" -o pertisk-chart ./cmd/server
	@echo "Build complete!"

# Run the server
run:
	@echo "Running pertisk-chart..."
	@go run ./cmd/server --debug --port=8080

# Run with custom port
run-port:
	@go run ./cmd/server --debug --port=$(PORT)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f pertisk-chart
	@rm -rf chartstorage
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run || echo "Install golangci-lint for linting: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run        - Run the server in debug mode"
	@echo "  run-port   - Run with custom port (PORT=8081 make run-port)"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code (requires golangci-lint)"
	@echo "  help       - Show this help message"

