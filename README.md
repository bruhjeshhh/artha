# Rent & Cost Analyzer - Terminal App

A terminal-only application for analyzing rent and living costs in Ashta, Madhya Pradesh, with AI-powered predictions and geospatial analysis.

## Features

- **AI-Powered Cost Prediction**: XGBoost regression model for personalized monthly cost predictions
- **Smart Rent Classification**: PyTorch DistilBERT NLP model to classify listings as "fair" or "overpriced"
- **Rent Analysis**: Scrape and analyze rental listings from NoBroker and OLX
- **Grocery Pricing**: Integrate with BigBasket and Blinkit APIs for grocery cost analysis
- **Transport Costs**: Fetch BCLL transport fares and calculate monthly commute costs
- **Inflation Tracking**: Load and visualize RBI and MP Government inflation data
- **Geospatial Analysis**: PostGIS-powered locality heatmaps, isochrone calculations, and nearby locality search
- **Cost Burden Index**: Calculate and visualize cost burden as percentage of income
- **Locality Comparison**: Compare costs across different localities
- **User Profiling**: Form-based user profile for personalized predictions

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15
- **Container**: Docker & Docker Compose

## Quick Setup

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose

### Installation

1. **Start PostgreSQL**:
```bash
docker-compose up -d
```

2. **Install Go dependencies**:
```bash
go mod download
```

3. **Run the application**:
```bash
go run main.go
```

The app will automatically:
- Connect to PostgreSQL
- Create necessary tables
- Seed mock data (rental listings, groceries, transport routes, inflation data)
- Present an interactive terminal menu

## Usage

### Main Menu Options

1. **Create User Profile** - Set up your personal details for customized predictions
2. **Analyze Rent Listings** - View AI-classified rental properties (fair vs overpriced)
3. **AI Cost Prediction** - Get XGBoost-powered monthly cost predictions
4. **Grocery Pricing** - View BigBasket and Blinkit grocery costs
5. **Transport Costs** - Calculate BCLL bus commute expenses
6. **Inflation Data** - View RBI and MP Government inflation trends
7. **Geospatial Analysis** - Heatmaps, isochrones, and nearby locality search
8. **Compare Localities** - Side-by-side cost comparison between areas
9. **Cost Burden Index** - See what percentage of income goes to living costs
10. **Exit**

### Sample Workflow
```
1. Create your user profile (Option 1)
   - Enter name, income, family size, preferred locality, commute distance

2. View rent listings with AI classification (Option 2)
   - See which properties are fair vs overpriced

3. Get AI cost prediction (Option 3)
   - XGBoost model predicts your total monthly costs

4. Check cost burden index (Option 9)
   - Understand affordability across different localities
```

## Database Schema

The app uses 4 main tables:

- **rental_listings**: Property data with AI classification
- **groceries**: Food items with prices from BigBasket/Blinkit
- **transport_routes**: BCLL bus routes and fares
- **inflation_data**: Historical inflation rates by category

## Mock Data

Since external APIs aren't available, the app generates realistic mock data:
- 20 rental listings across 6 localities in Ashta
- 8 common grocery items
- Transport routes between all localities
- 6 months of inflation data across 4 categories

## Demo Notes

This is a demo application with:
- Simulated AI models (XGBoost, DistilBERT) for supervisor demonstration
- Mock data that resembles real Ashta, MP data
- Minimal infrastructure (just Go + PostgreSQL)
- Terminal-only interface for simplicity

## Troubleshooting

**Connection refused to PostgreSQL**:
```bash
# Check if container is running
docker ps

# Restart if needed
docker-compose restart

# Check logs
docker-compose logs postgres
```

**Port 5432 already in use**:
```bash
# Change port in docker-compose.yml to 5433:5432
# Then update connection string in main.go
```

## License

MIT License - Educational/Demo purposes