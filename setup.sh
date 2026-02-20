#!/bin/bash



# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose not found. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Please install Go 1.21 or higher first."
    exit 1
fi

echo "✅ Docker found"
echo "✅ Docker Compose found"
echo "✅ Go found"
echo ""

# Start PostgreSQL only (services can be started with: make run-all)
echo "🐘 Starting PostgreSQL container..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

# Check if PostgreSQL is running
if docker ps | grep -q rentanalyzer-db; then
    echo "✅ PostgreSQL is running"
else
    echo "❌ Failed to start PostgreSQL"
    exit 1
fi

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║    Setup Complete! Ready to run the application.           ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Option A - Microservices (recommended):"
echo "  make run-all    # Start all services in Docker"
echo "  make run        # Run the CLI"
echo ""
echo "Option B - Local binaries (after make build):"
echo "  ./bin/user-service & ./bin/rental-service & ... (see Makefile)"
echo "  make run        # Run the CLI"
echo ""
echo "To stop later: docker-compose down"
echo ""