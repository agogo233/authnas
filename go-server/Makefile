.PHONY: build run test clean migrate

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
BINARY_NAME=authnas
BINARY_PATH=./bin/$(BINARY_NAME)

# Directories
CMD_DIR=./cmd/server
MIGRATIONS_DIR=./migrations

# Build the application
build:
	mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(CMD_DIR)

# Run the application
run: build
	$(BINARY_PATH)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	rm -rf bin
	$(GOCLEAN) -cache

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run database migrations
migrate:
	@echo "Running migrations..."
	@find $(MIGRATIONS_DIR) -name "*.sql" -exec echo "Executing {}" \; -exec cat {} \;

# Generate JWT keys
keys:
	@echo "Generating JWT keys..."
	@mkdir -p keys
	openssl genrsa -out keys/private.pem 2048
	openssl rsa -in keys/private.pem -pubout -out keys/certificate.pem

# Development helpers
dev-deps:
	$(GOMOD) download

# Docker commands
docker-build:
	docker build -t authnas .

docker-run:
	docker run -p 8080:8080 authnas

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

# Create data directory for SQLite
data-dir:
	mkdir -p data
