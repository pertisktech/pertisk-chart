.PHONY: build run run-dev clean test help install-air

# Build the server
build:
	@echo "Building pertisk-chart..."
	@go build -ldflags="-s -w" -o pertisk-chart ./cmd/server
	@echo "Build complete!"

# Run the server
run:
	@echo "Running pertisk-chart..."
	@go run ./cmd/server --debug --port=7080

# Run with hot reload using Air
run-dev:
	@echo "Running pertisk-chart with hot reload..."
	@if ! command -v air > /dev/null; then \
		echo "Air is not installed. Installing..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@air

# Run with custom port
run-port:
	@go run ./cmd/server --debug --port=$(PORT)

# Install Air for hot reload
install-air:
	@echo "Installing Air..."
	@go install github.com/air-verse/air@latest
	@echo "Air installed! Run 'make run-dev' to use hot reload."

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

# Create admin user
create-admin:
	@if [ -z "$(USERNAME)" ] || [ -z "$(EMAIL)" ] || [ -z "$(PASSWORD)" ]; then \
		echo "Error: USERNAME, EMAIL, and PASSWORD are required"; \
		echo "Usage: make create-admin USERNAME=admin EMAIL=admin@example.com PASSWORD=secret123"; \
		exit 1; \
	fi
	@go run cmd/create-admin/main.go \
		-username "$(USERNAME)" \
		-email "$(EMAIL)" \
		-password "$(PASSWORD)" \
		-db-type "$(or $(DB_TYPE),sqlite)" \
		-db-dsn "$(DB_DSN)" \
		-data-dir "$(or $(DATA_DIR),./data)"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run        - Run the server in debug mode"
	@echo "  run-dev    - Run with hot reload using Air"
	@echo "  run-port   - Run with custom port (PORT=8081 make run-port)"
	@echo "  install-air - Install Air for hot reload"
	@echo "  create-admin - Create an admin user (requires USERNAME, EMAIL, PASSWORD)"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code (requires golangci-lint)"
	@echo "  help       - Show this help message"

