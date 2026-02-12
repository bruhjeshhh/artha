.PHONY: setup run run-cli run-all build clean db-start db-stop db-logs help

help:
	@echo "Rent & Cost Analyzer (Microservices) - Available Commands:"
	@echo ""
	@echo "  make setup      - Set up database and dependencies"
	@echo "  make build      - Build all services and CLI"
	@echo "  make run        - Run CLI (requires services running)"
	@echo "  make run-cli    - Same as run"
	@echo "  make run-all    - Start all microservices via Docker Compose"
	@echo "  make db-start   - Start PostgreSQL only"
	@echo "  make db-stop    - Stop PostgreSQL"
	@echo "  make db-logs    - View PostgreSQL logs"
	@echo "  make clean      - Stop containers and clean up"
	@echo ""

setup:
	@echo "🚀 Setting up Rent & Cost Analyzer..."
	@./setup.sh

build:
	@echo "🔨 Building all services and CLI..."
	@go build -o bin/user-service ./cmd/user-service
	@go build -o bin/rental-service ./cmd/rental-service
	@go build -o bin/grocery-service ./cmd/grocery-service
	@go build -o bin/transport-service ./cmd/transport-service
	@go build -o bin/inflation-service ./cmd/inflation-service
	@go build -o bin/geospatial-service ./cmd/geospatial-service
	@go build -o bin/cost-prediction-service ./cmd/cost-prediction-service
	@go build -o bin/cli ./cmd/cli
	@echo "✅ Build complete. Binaries in ./bin/"

run run-cli:
	@echo "▶️  Starting CLI (ensure services are running: make run-all or run binaries in bin/)..."
	@go run ./cmd/cli

run-all:
	@echo "🐳 Starting all microservices with Docker Compose..."
	@docker-compose up -d
	@echo "✅ Services starting. Use 'make run' for CLI (SERVICES_HOST=localhost)."

db-start:
	@echo "🐘 Starting PostgreSQL..."
	@docker-compose up -d postgres
	@echo "✅ PostgreSQL started"

db-stop:
	@echo "🛑 Stopping PostgreSQL..."
	@docker-compose stop postgres
	@echo "✅ PostgreSQL stopped"

db-logs:
	@docker-compose logs -f postgres

clean:
	@echo "🧹 Cleaning up..."
	@docker-compose down -v
	@rm -rf bin/
	@echo "✅ Cleanup complete"
