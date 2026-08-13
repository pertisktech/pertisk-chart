.PHONY: build run run-dev clean test help install-air elete-tag create-tag retag clean-tag rpm rpm-native rpm-install

VERSION ?= $(patsubst v%,%,$(shell git describe --tags --abbrev=0 2>/dev/null || echo 0.1.2))
RELEASE ?= 1
ALMA_VERSION ?= 9

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
	@rm -f pertisk-chart pertisk-chart-create-admin
	@rm -rf chartstorage build/rpm
	@echo "Clean complete!"

# Build AlmaLinux/RHEL RPM via Docker (works on macOS and Linux)
rpm:
	@echo "Building pertisk-chart-$(VERSION)-$(RELEASE) RPM for AlmaLinux $(ALMA_VERSION)..."
	@mkdir -p dist
	docker buildx build \
		--output type=local,dest=dist \
		--build-arg VERSION=$(VERSION) \
		--build-arg RELEASE=$(RELEASE) \
		--build-arg ALMA_VERSION=$(ALMA_VERSION) \
		-f packaging/rpm/Dockerfile \
		.
	@echo ""
	@echo "RPMs written to dist/:"
	@ls -lh dist/*.rpm 2>/dev/null || ls -lh dist/

# Build RPM on the host (requires AlmaLinux/RHEL with rpmbuild and Go 1.21+)
rpm-native:
	@chmod +x packaging/rpm/build.sh packaging/pertisk-chart.sh
	VERSION=$(VERSION) RELEASE=$(RELEASE) ./packaging/rpm/build.sh

# Install the built RPM (AlmaLinux/RHEL only)
rpm-install:
	@rpm_file=$$(ls -1t dist/pertisk-chart-$(VERSION)-$(RELEASE).*.rpm 2>/dev/null | grep -v src.rpm | head -1); \
	if [ -z "$$rpm_file" ]; then \
		echo "No RPM found. Run 'make rpm' first."; \
		exit 1; \
	fi; \
	if [ -f /etc/almalinux-release ] || [ -f /etc/redhat-release ]; then \
		echo "Installing $$rpm_file ..."; \
		sudo dnf install -y "$$rpm_file"; \
	else \
		echo "This host is not AlmaLinux/RHEL. Copy the RPM to a target host:"; \
		echo "  $$rpm_file"; \
		echo ""; \
		echo "On AlmaLinux:"; \
		echo "  sudo dnf install -y $$(basename $$rpm_file)"; \
		echo "  sudo pertisk-chart-create-admin -username admin -email admin@example.com -password CHANGE_ME -data-dir /var/lib/pertisk-chart -db-type sqlite"; \
		echo "  sudo systemctl enable --now pertisk-chart"; \
	fi

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
	@echo "  rpm        - Build AlmaLinux RPM via Docker (VERSION=0.1.2 ALMA_VERSION=9)"
	@echo "  rpm-native - Build RPM on this host (requires rpmbuild and Go 1.21+)"
	@echo "  rpm-install - Install the built RPM (AlmaLinux/RHEL) or print install steps"
	@echo "  test       - Run tests"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code (requires golangci-lint)"
	@echo "  help       - Show this help message"

delete-tag:
ifndef TAG
	$(error TAG is not set. Usage: make delete-tag TAG=v1.0.0)
endif
	@echo "Deleting tag $(TAG)..."
	git tag -d $(TAG)
	git push origin -d $(TAG)

# Create a new tag
create-tag:
ifndef TAG
	$(error TAG is not set. Usage: make create-tag TAG=v1.0.0)
endif
	@echo "Creating tag $(TAG)..."
	git tag $(TAG)
	git push origin $(TAG)

# Delete and recreate a tag (force update)
# Works even if tag doesn't exist yet (creates new tag)
retag:
ifndef TAG
	$(error TAG is not set. Usage: make retag TAG=v1.0.0)
endif
	@echo "Recreating tag $(TAG)..."
	@echo "Deleting local tag (if exists)..."
	-git tag -d $(TAG) 2>/dev/null || true
	@echo "Deleting remote tag (if exists)..."
	-git push origin -d $(TAG) 2>/dev/null || true
	@echo "Creating new tag $(TAG)..."
	git tag $(TAG)
	@echo "Pushing tag $(TAG) to origin..."
	git push origin $(TAG)
	@echo "✓ Tag $(TAG) created and pushed successfully"

# Alias for retag
clean-tag: retag