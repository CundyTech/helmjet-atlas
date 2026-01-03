.PHONY: help build run test clean install-deps docker-build docker-run docker-stop docker-logs

help:
	@echo "Available targets:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application locally"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install-deps   - Download Go dependencies"
	@echo "  make test           - Run tests (placeholder)"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run with Docker Compose"
	@echo "  make docker-stop    - Stop Docker Compose"
	@echo "  make docker-logs    - View Docker Compose logs"
	@echo "  make format         - Format code with gofmt"
	@echo "  make lint           - Run linter (if available)"

build:
	@echo "Building helmjet-atlas..."
	go build -o helmjet-atlas .

run:
	@echo "Running helmjet-atlas..."
	@echo "Make sure MongoDB is running!"
	@MONGO_URI="mongodb://localhost:27017" \
	MONGO_DB="helmjet-atlas" \
	PORT="8080" \
	go run main.go

clean:
	@echo "Cleaning up..."
	rm -f helmjet-atlas

install-deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

test:
	@echo "Running tests..."
	go test ./...

docker-build:
	@echo "Building Docker image..."
	docker build -t helmjet-atlas:latest .

docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build -d

docker-stop:
	@echo "Stopping services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

docker-clean:
	@echo "Removing containers and volumes..."
	docker-compose down -v

format:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Development setup
dev-setup: install-deps
	@echo "Development environment ready!"
	@echo "Start MongoDB with: docker run -d -p 27017:27017 mongo:latest"
	@echo "Then run: make run"

# Production build
prod-build: clean
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o helmjet-atlas .
	@echo "Binary ready: ./helmjet-atlas"
