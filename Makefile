.PHONY: setup run clean db-start db-stop db-logs help

help:
	@echo "Rent & Cost Analyzer - Available Commands:"
	@echo ""
	@echo "  make setup      - Set up database and dependencies"
	@echo "  make run        - Run the application"
	@echo "  make db-start   - Start PostgreSQL container"
	@echo "  make db-stop    - Stop PostgreSQL container"
	@echo "  make db-logs    - View PostgreSQL logs"
	@echo "  make clean      - Stop containers and clean up"
	@echo ""

setup:
	@echo "🚀 Setting up Rent & Cost Analyzer..."
	@./setup.sh

run:
	@echo "▶️  Starting Rent & Cost Analyzer..."
	@go run main.go

db-start:
	@echo "🐘 Starting PostgreSQL..."
	@docker-compose up -d
	@echo "✅ PostgreSQL started"

db-stop:
	@echo "🛑 Stopping PostgreSQL..."
	@docker-compose stop
	@echo "✅ PostgreSQL stopped"

db-logs:
	@docker-compose logs -f postgres

clean:
	@echo "🧹 Cleaning up..."
	@docker-compose down -v
	@echo "✅ Cleanup complete"