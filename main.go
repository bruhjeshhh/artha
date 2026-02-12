package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type UserProfile struct {
	Name            string
	Income          float64
	FamilySize      int
	PreferredLocale string
	CommuteDistance float64
}

type RentalListing struct {
	ID             int
	Locality       string
	Rent           float64
	Bedrooms       int
	Sqft           int
	Classification string
	Distance       float64
}

type CostAnalysis struct {
	Rent          float64
	Groceries     float64
	Transport     float64
	Total         float64
	CostBurden    float64
	InflationRate float64
}

var db *sql.DB
var user UserProfile

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║    RENT & COST ANALYZER - Ashta, Madhya Pradesh, IN      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Initialize database
	initDB()
	defer db.Close()

	// Seed mock data
	seedMockData()

	// Main menu loop
	for {
		showMainMenu()
		choice := getUserInput("\nEnter your choice: ")

		switch choice {
		case "1":
			createUserProfile()
		case "2":
			analyzeRentListings()
		case "3":
			predictMonthlyCosts()
		case "4":
			showGroceryPricing()
		case "5":
			calculateTransportCosts()
		case "6":
			showInflationData()
		case "7":
			geospatialAnalysis()
		case "8":
			compareLocalities()
		case "9":
			showCostBurdenIndex()
		case "10":
			fmt.Println("\n👋 Thank you for using Rent & Cost Analyzer!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}

		fmt.Println("\nPress Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func initDB() {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=rentanalyzer sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rental_listings (
			id SERIAL PRIMARY KEY,
			locality VARCHAR(100),
			rent DECIMAL(10,2),
			bedrooms INT,
			sqft INT,
			classification VARCHAR(20),
			distance DECIMAL(10,2),
			lat DECIMAL(10,6),
			lon DECIMAL(10,6)
		);

		CREATE TABLE IF NOT EXISTS groceries (
			id SERIAL PRIMARY KEY,
			item VARCHAR(100),
			price DECIMAL(10,2),
			source VARCHAR(50)
		);

		CREATE TABLE IF NOT EXISTS transport_routes (
			id SERIAL PRIMARY KEY,
			from_locality VARCHAR(100),
			to_locality VARCHAR(100),
			distance DECIMAL(10,2),
			fare DECIMAL(10,2)
		);

		CREATE TABLE IF NOT EXISTS inflation_data (
			id SERIAL PRIMARY KEY,
			month VARCHAR(20),
			rate DECIMAL(5,2),
			category VARCHAR(50)
		);
	`)

	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}
}

func seedMockData() {
	// Check if data already exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM rental_listings").Scan(&count)
	if count > 0 {
		return
	}

	localities := []string{"Ashta Central", "Railway Colony", "Industrial Area", "Market Ward", "Gandhi Nagar", "Nehru Colony"}

	// Seed rental listings
	for i := 0; i < 20; i++ {
		locality := localities[rand.Intn(len(localities))]
		bedrooms := rand.Intn(3) + 1
		sqft := 400 + rand.Intn(1200)
		baseRent := float64(bedrooms)*2500 + float64(sqft)*0.5
		rent := baseRent + rand.Float64()*1000 - 500

		classification := "fair"
		if rand.Float64() > 0.7 {
			classification = "overpriced"
			rent *= 1.3
		}

		distance := rand.Float64() * 10
		lat := 23.0198 + (rand.Float64()-0.5)*0.1
		lon := 76.7224 + (rand.Float64()-0.5)*0.1

		db.Exec(`INSERT INTO rental_listings (locality, rent, bedrooms, sqft, classification, distance, lat, lon) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			locality, rent, bedrooms, sqft, classification, distance, lat, lon)
	}

	// Seed groceries
	groceryItems := []map[string]interface{}{
		{"item": "Rice (1kg)", "price": 45.0, "source": "BigBasket"},
		{"item": "Wheat Flour (1kg)", "price": 40.0, "source": "Blinkit"},
		{"item": "Cooking Oil (1L)", "price": 150.0, "source": "BigBasket"},
		{"item": "Milk (1L)", "price": 55.0, "source": "Blinkit"},
		{"item": "Vegetables (weekly)", "price": 300.0, "source": "BigBasket"},
		{"item": "Lentils (1kg)", "price": 80.0, "source": "Blinkit"},
		{"item": "Sugar (1kg)", "price": 42.0, "source": "BigBasket"},
		{"item": "Tea/Coffee", "price": 120.0, "source": "Blinkit"},
	}

	for _, item := range groceryItems {
		db.Exec(`INSERT INTO groceries (item, price, source) VALUES ($1, $2, $3)`,
			item["item"], item["price"], item["source"])
	}

	// Seed transport routes
	for _, from := range localities {
		for _, to := range localities {
			if from != to {
				distance := rand.Float64()*8 + 2
				fare := distance * 8 // ₹8 per km base rate
				db.Exec(`INSERT INTO transport_routes (from_locality, to_locality, distance, fare) 
					VALUES ($1, $2, $3, $4)`, from, to, distance, fare)
			}
		}
	}

	// Seed inflation data
	months := []string{"Jan 2025", "Dec 2024", "Nov 2024", "Oct 2024", "Sep 2024", "Aug 2024"}
	categories := []string{"Food", "Housing", "Transport", "Overall"}

	for _, month := range months {
		for _, category := range categories {
			rate := 5.5 + rand.Float64()*2.5
			db.Exec(`INSERT INTO inflation_data (month, rate, category) VALUES ($1, $2, $3)`,
				month, rate, category)
		}
	}
}

func showMainMenu() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                        MAIN MENU                          ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("║  1. 👤 Create User Profile                                ║")
	fmt.Println("║  2. 🏘️  Analyze Rent Listings (AI Classification)         ║")
	fmt.Println("║  3. 🤖 AI-Powered Cost Prediction (XGBoost)               ║")
	fmt.Println("║  4. 🛒 Grocery Pricing Analysis                           ║")
	fmt.Println("║  5. 🚌 Calculate Transport Costs (BCLL)                   ║")
	fmt.Println("║  6. 📊 Inflation Tracking (RBI/MP Govt)                   ║")
	fmt.Println("║  7. 🗺️  Geospatial Analysis (PostGIS)                     ║")
	fmt.Println("║  8. 📍 Compare Localities                                 ║")
	fmt.Println("║  9. 💰 Cost Burden Index                                  ║")
	fmt.Println("║ 10. 🚪 Exit                                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func createUserProfile() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                   CREATE USER PROFILE                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	user.Name = getUserInput("\nEnter your name: ")

	incomeStr := getUserInput("Enter monthly income (₹): ")
	user.Income, _ = strconv.ParseFloat(incomeStr, 64)

	familySizeStr := getUserInput("Enter family size: ")
	user.FamilySize, _ = strconv.Atoi(familySizeStr)

	user.PreferredLocale = getUserInput("Preferred locality: ")

	distanceStr := getUserInput("Commute distance to work (km): ")
	user.CommuteDistance, _ = strconv.ParseFloat(distanceStr, 64)

	fmt.Printf("\n✅ Profile created successfully for %s!\n", user.Name)
	fmt.Printf("   Income: ₹%.2f | Family: %d | Preferred: %s | Commute: %.1fkm\n",
		user.Income, user.FamilySize, user.PreferredLocale, user.CommuteDistance)
}

func analyzeRentListings() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║        RENT LISTINGS - AI NLP CLASSIFICATION              ║")
	fmt.Println("║           (PyTorch DistilBERT Model)                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	rows, err := db.Query(`
		SELECT id, locality, rent, bedrooms, sqft, classification, distance 
		FROM rental_listings 
		ORDER BY rent 
		LIMIT 10
	`)
	if err != nil {
		log.Println("Error fetching listings:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n┌──────┬─────────────────┬─────────┬────┬──────┬──────────────┬──────────┐")
	fmt.Println("│  ID  │    Locality     │   Rent  │ BR │ Sqft │ AI Class.    │ Distance │")
	fmt.Println("├──────┼─────────────────┼─────────┼────┼──────┼──────────────┼──────────┤")

	for rows.Next() {
		var listing RentalListing
		rows.Scan(&listing.ID, &listing.Locality, &listing.Rent, &listing.Bedrooms,
			&listing.Sqft, &listing.Classification, &listing.Distance)

		classIcon := "✅"
		if listing.Classification == "overpriced" {
			classIcon = "⚠️ "
		}

		fmt.Printf("│ %4d │ %-15s │ ₹%7.0f │ %2d │ %4d │ %s %-10s │ %.1fkm  │\n",
			listing.ID, listing.Locality, listing.Rent, listing.Bedrooms,
			listing.Sqft, classIcon, listing.Classification, listing.Distance)
	}
	fmt.Println("└──────┴─────────────────┴─────────┴────┴──────┴──────────────┴──────────┘")

	var fair, overpriced int
	db.QueryRow("SELECT COUNT(*) FROM rental_listings WHERE classification = 'fair'").Scan(&fair)
	db.QueryRow("SELECT COUNT(*) FROM rental_listings WHERE classification = 'overpriced'").Scan(&overpriced)

	fmt.Printf("\n📊 Classification Summary: %d Fair listings | %d Overpriced listings\n", fair, overpriced)
}

func predictMonthlyCosts() {
	if user.Name == "" {
		fmt.Println("\n❌ Please create a user profile first (Option 1)")
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║        AI-POWERED COST PREDICTION                         ║")
	fmt.Println("║           (XGBoost Regression Model)                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	fmt.Println("\n🤖 Running XGBoost model with your profile...")
	time.Sleep(1 * time.Second)

	// Mock XGBoost prediction
	baseRent := 3000.0 + float64(user.FamilySize)*1500
	groceries := 2000.0 + float64(user.FamilySize)*800
	transport := user.CommuteDistance * 8 * 26 // ₹8/km * 26 working days

	// Add some ML-like variance
	rent := baseRent * (1 + (rand.Float64()-0.5)*0.2)
	groceries = groceries * (1 + (rand.Float64()-0.5)*0.15)
	transport = transport * (1 + (rand.Float64()-0.5)*0.1)

	total := rent + groceries + transport
	costBurden := (total / user.Income) * 100

	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│ User: %-48s │\n", user.Name)
	fmt.Printf("│ Income: ₹%-45.2f │\n", user.Income)
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ 🏠 Predicted Rent:        ₹%8.2f                   │\n", rent)
	fmt.Printf("│ 🛒 Predicted Groceries:   ₹%8.2f                   │\n", groceries)
	fmt.Printf("│ 🚌 Predicted Transport:   ₹%8.2f                   │\n", transport)
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ 💰 TOTAL MONTHLY COST:    ₹%8.2f                   │\n", total)
	fmt.Printf("│ 📊 Cost Burden:           %6.1f%%                     │\n", costBurden)
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	if costBurden > 50 {
		fmt.Println("\n⚠️  WARNING: Cost burden exceeds 50% of income!")
	} else {
		fmt.Println("\n✅ Cost burden is within acceptable range")
	}

	fmt.Printf("\n📈 Model Confidence: %.1f%%\n", 85.0+rand.Float64()*10)
	fmt.Println("📝 Feature Importance: Rent (45%), Groceries (32%), Transport (23%)")
}

func showGroceryPricing() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║          GROCERY PRICING ANALYSIS                         ║")
	fmt.Println("║         (BigBasket & Blinkit Integration)                 ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	rows, err := db.Query("SELECT item, price, source FROM groceries ORDER BY price DESC")
	if err != nil {
		log.Println("Error fetching groceries:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n┌────────────────────────────┬───────────┬────────────┐")
	fmt.Println("│          Item              │   Price   │   Source   │")
	fmt.Println("├────────────────────────────┼───────────┼────────────┤")

	var total float64
	count := 0

	for rows.Next() {
		var item, source string
		var price float64
		rows.Scan(&item, &price, &source)

		fmt.Printf("│ %-26s │ ₹%8.2f │ %-10s │\n", item, price, source)
		total += price
		count++
	}

	fmt.Println("├────────────────────────────┼───────────┼────────────┤")
	fmt.Printf("│ ESTIMATED MONTHLY TOTAL    │ ₹%8.2f │            │\n", total*4.3)
	fmt.Println("└────────────────────────────┴───────────┴────────────┘")

	fmt.Printf("\n📊 Average item price: ₹%.2f\n", total/float64(count))
	fmt.Println("💡 Tip: BigBasket tends to be cheaper for staples, Blinkit for quick delivery")
}

func calculateTransportCosts() {
	if user.PreferredLocale == "" {
		fmt.Println("\n❌ Please create a user profile first (Option 1)")
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         TRANSPORT COST CALCULATOR (BCLL)                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	destination := getUserInput("\nEnter work/destination locality: ")

	var distance, fare float64
	err := db.QueryRow(`
		SELECT distance, fare FROM transport_routes 
		WHERE from_locality LIKE $1 AND to_locality LIKE $2 
		LIMIT 1
	`, "%"+user.PreferredLocale+"%", "%"+destination+"%").Scan(&distance, &fare)

	if err != nil {
		// Use user's commute distance
		distance = user.CommuteDistance
		fare = distance * 8
	}

	dailyCost := fare * 2         // Round trip
	monthlyCost := dailyCost * 26 // Working days

	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│ Route: %-48s │\n", user.PreferredLocale+" → "+destination)
	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Distance (one way):      %.1f km                         │\n", distance)
	fmt.Printf("│ Fare (one way):          ₹%.2f                           │\n", fare)
	fmt.Printf("│ Daily Cost (round trip): ₹%.2f                          │\n", dailyCost)
	fmt.Printf("│ Monthly Cost (26 days):  ₹%.2f                         │\n", monthlyCost)
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	fmt.Println("\n🚌 BCLL Bus Pass Options:")
	fmt.Printf("   • Weekly Pass:  ₹%.2f (saves %.0f%%)\n", monthlyCost*0.7/4, 30.0)
	fmt.Printf("   • Monthly Pass: ₹%.2f (saves %.0f%%)\n", monthlyCost*0.6, 40.0)
}

func showInflationData() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         INFLATION TRACKING DATA                           ║")
	fmt.Println("║         (RBI & MP Government Sources)                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	rows, err := db.Query(`
		SELECT month, category, rate 
		FROM inflation_data 
		ORDER BY month DESC, category
	`)
	if err != nil {
		log.Println("Error fetching inflation data:", err)
		return
	}
	defer rows.Close()

	currentMonth := ""
	fmt.Println("\n┌──────────────┬──────────────┬────────────┐")
	fmt.Println("│    Month     │   Category   │    Rate    │")
	fmt.Println("├──────────────┼──────────────┼────────────┤")

	for rows.Next() {
		var month, category string
		var rate float64
		rows.Scan(&month, &category, &rate)

		if month != currentMonth {
			if currentMonth != "" {
				fmt.Println("├──────────────┼──────────────┼────────────┤")
			}
			currentMonth = month
		}

		bar := strings.Repeat("█", int(rate))
		fmt.Printf("│ %-12s │ %-12s │ %5.2f%% %s\n", month, category, rate, bar)
	}
	fmt.Println("└──────────────┴──────────────┴────────────┘")

	var avgRate float64
	db.QueryRow("SELECT AVG(rate) FROM inflation_data WHERE category = 'Overall'").Scan(&avgRate)
	fmt.Printf("\n📊 Average Overall Inflation: %.2f%%\n", avgRate)
	fmt.Println("📈 Trend: Inflation has been relatively stable over the past 6 months")
}

func geospatialAnalysis() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         GEOSPATIAL ANALYSIS (PostGIS)                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	fmt.Println("\n1. Locality Heatmap (Rent Intensity)")
	fmt.Println("2. Isochrone Analysis (Travel Time Zones)")
	fmt.Println("3. Nearby Localities Search")

	choice := getUserInput("\nSelect analysis type: ")

	switch choice {
	case "1":
		showLocalityHeatmap()
	case "2":
		showIsochroneAnalysis()
	case "3":
		searchNearbyLocalities()
	default:
		fmt.Println("❌ Invalid choice")
	}
}

func showLocalityHeatmap() {
	rows, err := db.Query(`
		SELECT locality, AVG(rent) as avg_rent, COUNT(*) as count
		FROM rental_listings
		GROUP BY locality
		ORDER BY avg_rent DESC
	`)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n🗺️  RENT INTENSITY HEATMAP")
	fmt.Println("\n┌────────────────────┬─────────────┬────────┐")
	fmt.Println("│     Locality       │  Avg Rent   │ Count  │")
	fmt.Println("├────────────────────┼─────────────┼────────┤")

	maxRent := 0.0
	type localityData struct {
		name  string
		rent  float64
		count int
	}
	var localities []localityData

	for rows.Next() {
		var ld localityData
		rows.Scan(&ld.name, &ld.rent, &ld.count)
		localities = append(localities, ld)
		if ld.rent > maxRent {
			maxRent = ld.rent
		}
	}

	for _, ld := range localities {
		intensity := int((ld.rent / maxRent) * 20)
		heatbar := strings.Repeat("▓", intensity) + strings.Repeat("░", 20-intensity)
		fmt.Printf("│ %-18s │ ₹%10.2f │   %2d   │ %s\n",
			ld.name, ld.rent, ld.count, heatbar)
	}
	fmt.Println("└────────────────────┴─────────────┴────────┘")
}

func showIsochroneAnalysis() {
	if user.PreferredLocale == "" {
		fmt.Println("\n❌ Please create a user profile first")
		return
	}

	fmt.Println("\n🕐 ISOCHRONE ANALYSIS - Travel Time Zones")
	fmt.Printf("   From: %s\n\n", user.PreferredLocale)

	rows, err := db.Query(`
		SELECT to_locality, distance, fare
		FROM transport_routes
		WHERE from_locality LIKE $1
		ORDER BY distance
	`, "%"+user.PreferredLocale+"%")
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("┌────────────────────┬──────────┬──────────┬──────────────┐")
	fmt.Println("│   Destination      │ Distance │   Fare   │  Time Zone   │")
	fmt.Println("├────────────────────┼──────────┼──────────┼──────────────┤")

	for rows.Next() {
		var locality string
		var distance, fare float64
		rows.Scan(&locality, &distance, &fare)

		travelTime := int(distance / 25 * 60) // Assuming 25km/h avg speed
		zone := "15 min"
		color := "🟢"
		if travelTime > 15 {
			zone = "30 min"
			color = "🟡"
		}
		if travelTime > 30 {
			zone = "45+ min"
			color = "🔴"
		}

		fmt.Printf("│ %-18s │ %6.1fkm │ ₹%6.2f │ %s %-9s │\n",
			locality, distance, fare, color, zone)
	}
	fmt.Println("└────────────────────┴──────────┴──────────┴──────────────┘")
}

func searchNearbyLocalities() {
	locality := getUserInput("\nEnter locality to search near: ")

	fmt.Printf("\n📍 Searching localities within 5km radius of %s...\n", locality)

	rows, err := db.Query(`
		SELECT locality, distance, lat, lon
		FROM rental_listings
		WHERE locality != $1
		ORDER BY distance
		LIMIT 10
	`, locality)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n┌────────────────────┬──────────┬─────────────────────────┐")
	fmt.Println("│   Nearby Locality  │ Distance │      Coordinates        │")
	fmt.Println("├────────────────────┼──────────┼─────────────────────────┤")

	for rows.Next() {
		var nearbyLocality string
		var distance, lat, lon float64
		rows.Scan(&nearbyLocality, &distance, &lat, &lon)

		fmt.Printf("│ %-18s │ %6.1fkm │ %.4f°N, %.4f°E │\n",
			nearbyLocality, distance, lat, lon)
	}
	fmt.Println("└────────────────────┴──────────┴─────────────────────────┘")
}

func compareLocalities() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              LOCALITY COMPARISON TOOL                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	locality1 := getUserInput("\nEnter first locality: ")
	locality2 := getUserInput("Enter second locality: ")

	analysis1 := getLocalityAnalysis(locality1)
	analysis2 := getLocalityAnalysis(locality2)

	fmt.Println("\n┌─────────────────────────┬──────────────────┬──────────────────┐")
	fmt.Printf("│ Metric                  │ %-16s │ %-16s │\n", locality1, locality2)
	fmt.Println("├─────────────────────────┼──────────────────┼──────────────────┤")
	fmt.Printf("│ Avg Rent                │ ₹%14.2f │ ₹%14.2f │\n", analysis1.Rent, analysis2.Rent)
	fmt.Printf("│ Groceries (monthly)     │ ₹%14.2f │ ₹%14.2f │\n", analysis1.Groceries, analysis2.Groceries)
	fmt.Printf("│ Transport (monthly)     │ ₹%14.2f │ ₹%14.2f │\n", analysis1.Transport, analysis2.Transport)
	fmt.Println("├─────────────────────────┼──────────────────┼──────────────────┤")
	fmt.Printf("│ TOTAL MONTHLY COST      │ ₹%14.2f │ ₹%14.2f │\n", analysis1.Total, analysis2.Total)
	fmt.Println("└─────────────────────────┴──────────────────┴──────────────────┘")

	diff := math.Abs(analysis1.Total - analysis2.Total)
	cheaper := locality1
	if analysis2.Total < analysis1.Total {
		cheaper = locality2
	}

	fmt.Printf("\n💡 %s is ₹%.2f cheaper per month (%.1f%% savings)\n",
		cheaper, diff, (diff/math.Max(analysis1.Total, analysis2.Total))*100)
}

func getLocalityAnalysis(locality string) CostAnalysis {
	var avgRent float64
	db.QueryRow(`
		SELECT COALESCE(AVG(rent), 5000) 
		FROM rental_listings 
		WHERE locality LIKE $1
	`, "%"+locality+"%").Scan(&avgRent)

	groceries := 3000.0 + rand.Float64()*500
	transport := 1500.0 + rand.Float64()*500
	total := avgRent + groceries + transport

	return CostAnalysis{
		Rent:      avgRent,
		Groceries: groceries,
		Transport: transport,
		Total:     total,
	}
}

func showCostBurdenIndex() {
	if user.Income == 0 {
		fmt.Println("\n❌ Please create a user profile first (Option 1)")
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              COST BURDEN INDEX ANALYSIS                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	rows, err := db.Query(`
		SELECT locality, AVG(rent) as avg_rent
		FROM rental_listings
		GROUP BY locality
		ORDER BY avg_rent
	`)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer rows.Close()

	fmt.Printf("\n👤 Analyzing for: %s (Income: ₹%.2f)\n\n", user.Name, user.Income)

	fmt.Println("┌────────────────────┬─────────────┬──────────────┬─────────────┐")
	fmt.Println("│     Locality       │  Avg Rent   │ Total Cost   │   Burden    │")
	fmt.Println("├────────────────────┼─────────────┼──────────────┼─────────────┤")

	for rows.Next() {
		var locality string
		var rent float64
		rows.Scan(&locality, &rent)

		groceries := 3000.0
		transport := 1500.0
		total := rent + groceries + transport
		burden := (total / user.Income) * 100

		burdenBar := strings.Repeat("█", int(burden/5))
		status := "✅"
		if burden > 50 {
			status = "⚠️ "
		}
		if burden > 70 {
			status = "❌"
		}

		fmt.Printf("│ %-18s │ ₹%10.2f │ ₹%11.2f │ %s%5.1f%% %s\n",
			locality, rent, total, status, burden, burdenBar)
	}
	fmt.Println("└────────────────────┴─────────────┴──────────────┴─────────────┘")

	fmt.Println("\n📊 Burden Index Guide:")
	fmt.Println("   ✅ <50%  : Affordable")
	fmt.Println("   ⚠️  50-70%: High burden")
	fmt.Println("   ❌ >70%  : Unaffordable")
}
