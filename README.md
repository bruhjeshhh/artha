# Rent & Cost Analyzer - Microservices

A terminal application for analyzing rent and living costs in Ashta, Madhya Pradesh, with AI-powered predictions and geospatial analysis. The application is split into **microservices** plus a **CLI client**.

## Architecture

| Service | Port | Responsibility |
|---------|------|----------------|
| **user-service** | 8081 | User profile (create/get) |
| **rental-service** | 8082 | Rent listings, AI classification summary, locality comparison, cost burden |
| **grocery-service** | 8083 | Grocery pricing (BigBasket/Blinkit style) |
| **transport-service** | 8084 | Transport routes, BCLL fares, isochrone |
| **inflation-service** | 8085 | Inflation data (RBI/MP Govt style) |
| **geospatial-service** | 8086 | Heatmap, nearby localities (PostGIS-style) |
| **cost-prediction-service** | 8087 | XGBoost-style cost prediction (stateless) |
| **CLI** | - | Terminal UI that calls all services |

All data services share a single PostgreSQL database (same schema as before); each service owns its tables and seeds its own data on first run.

## Features

- **AI-Powered Cost Prediction**: XGBoost-style regression for personalized monthly cost predictions
- **Smart Rent Classification**: Listings classified as "fair" or "overpriced"
- **Rent Analysis**: Rental listings with locality and distance
- **Grocery Pricing**: BigBasket/Blinkit-style grocery cost analysis
- **Transport Costs**: BCLL-style transport fares and commute costs
- **Inflation Tracking**: RBI/MP Government-style inflation data
- **Geospatial Analysis**: Locality heatmaps, isochrones, nearby locality search
- **Cost Burden Index**: Cost burden as percentage of income
- **Locality Comparison**: Compare costs across localities
- **User Profiling**: Profile for personalized predictions

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15
- **Container**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose

### 1. Setup (PostgreSQL + deps)

```bash
make setup
# or: ./setup.sh
```

### 2. Run all microservices

```bash
make run-all
# Starts: postgres, user, rental, grocery, transport, inflation, geospatial, cost-prediction
```

### 3. Run the CLI

```bash
make run
# or: go run ./cmd/cli
```

The CLI talks to `localhost:8081-8087` by default. Override with:

```bash
SERVICES_HOST=localhost make run
```

## Local development (binaries)

Build and run services locally against a running Postgres:

```bash
make db-start
make build

# In separate terminals (or background):
./bin/user-service &
./bin/rental-service &
./bin/grocery-service &
./bin/transport-service &
./bin/inflation-service &
./bin/geospatial-service &
./bin/cost-prediction-service &

make run
```

## Main Menu (CLI)

1. **Create User Profile** – Name, income, family size, preferred locality, commute
2. **Analyze Rent Listings** – AI-classified listings (fair/overpriced)
3. **AI Cost Prediction** – XGBoost-style monthly cost prediction
4. **Grocery Pricing** – Grocery items and monthly estimate
5. **Transport Costs** – BCLL-style route and monthly cost
6. **Inflation Data** – Inflation by month/category
7. **Geospatial Analysis** – Heatmap, isochrone, nearby localities
8. **Compare Localities** – Side-by-side cost comparison
9. **Cost Burden Index** – Burden % by locality
10. **Exit**

## Database

- **rental_listings** – rental-service
- **groceries** – grocery-service
- **transport_routes** – transport-service
- **inflation_data** – inflation-service
- **users** – user-service

Geospatial service reads `rental_listings` (read-only). Cost-prediction service is stateless and uses the user profile from the CLI request.

## Makefile

- `make help` – List commands
- `make setup` – Setup DB and deps
- `make build` – Build all binaries to `bin/`
- `make run` / `make run-cli` – Run CLI
- `make run-all` – Start all services with Docker Compose
- `make db-start` / `make db-stop` / `make db-logs` – Postgres only
- `make clean` – Stop containers, remove `bin/`

## License

MIT License – Educational/Demo purposes.
