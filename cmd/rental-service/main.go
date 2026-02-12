package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"rent-cost-analyzer/internal/db"
	"rent-cost-analyzer/pkg/models"
)

var conn *sql.DB

func main() {
	c, err := db.Open()
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer c.Close()
	conn = c

	initTables(conn)
	seedMockData(conn)

	http.HandleFunc("/listings", handleListings)
	http.HandleFunc("/listings/summary", handleListingsSummary)
	http.HandleFunc("/compare", handleCompare)
	http.HandleFunc("/cost-burden", handleCostBurden)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("rental-service listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func initTables(c *sql.DB) {
	_, err := c.Exec(`
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
	`)
	if err != nil {
		log.Fatal("create table:", err)
	}
}

func seedMockData(c *sql.DB) {
	var count int
	c.QueryRow("SELECT COUNT(*) FROM rental_listings").Scan(&count)
	if count > 0 {
		return
	}

	localities := []string{"Ashta Central", "Railway Colony", "Industrial Area", "Market Ward", "Gandhi Nagar", "Nehru Colony"}

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

		c.Exec(`INSERT INTO rental_listings (locality, rent, bedrooms, sqft, classification, distance, lat, lon) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			locality, rent, bedrooms, sqft, classification, distance, lat, lon)
	}
}

func handleListings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	rows, err := conn.Query(`
		SELECT id, locality, rent, bedrooms, sqft, classification, distance 
		FROM rental_listings 
		ORDER BY rent 
		LIMIT 10
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []models.RentalListing
	for rows.Next() {
		var l models.RentalListing
		err := rows.Scan(&l.ID, &l.Locality, &l.Rent, &l.Bedrooms, &l.Sqft, &l.Classification, &l.Distance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, l)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"listings": list})
}

func handleListingsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var fair, overpriced int
	conn.QueryRow("SELECT COUNT(*) FROM rental_listings WHERE classification = 'fair'").Scan(&fair)
	conn.QueryRow("SELECT COUNT(*) FROM rental_listings WHERE classification = 'overpriced'").Scan(&overpriced)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"fair":       fair,
		"overpriced": overpriced,
	})
}

func handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	loc1 := r.URL.Query().Get("loc1")
	loc2 := r.URL.Query().Get("loc2")
	if loc1 == "" || loc2 == "" {
		http.Error(w, "loc1 and loc2 required", http.StatusBadRequest)
		return
	}

	a1 := getLocalityAnalysis(conn, loc1)
	a2 := getLocalityAnalysis(conn, loc2)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"locality1": loc1,
		"locality2": loc2,
		"analysis1": a1,
		"analysis2": a2,
	})
}

func getLocalityAnalysis(c *sql.DB, locality string) models.CostAnalysis {
	var avgRent float64
	c.QueryRow(`
		SELECT COALESCE(AVG(rent), 5000) 
		FROM rental_listings 
		WHERE locality LIKE $1
	`, "%"+locality+"%").Scan(&avgRent)

	groceries := 3000.0 + rand.Float64()*500
	transport := 1500.0 + rand.Float64()*500
	total := avgRent + groceries + transport

	return models.CostAnalysis{
		Rent:      avgRent,
		Groceries: groceries,
		Transport: transport,
		Total:     total,
	}
}

func handleCostBurden(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	incomeStr := r.URL.Query().Get("income")
	if incomeStr == "" {
		http.Error(w, "income required", http.StatusBadRequest)
		return
	}
	income, err := strconv.ParseFloat(incomeStr, 64)
	if err != nil || income <= 0 {
		http.Error(w, "invalid income", http.StatusBadRequest)
		return
	}

	rows, err := conn.Query(`
		SELECT locality, AVG(rent) as avg_rent
		FROM rental_listings
		GROUP BY locality
		ORDER BY avg_rent
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type burdenRow struct {
		Locality string  `json:"locality"`
		AvgRent  float64 `json:"avg_rent"`
		Total    float64 `json:"total"`
		Burden   float64 `json:"burden_pct"`
	}
	var result []burdenRow

	for rows.Next() {
		var locality string
		var rent float64
		rows.Scan(&locality, &rent)

		groceries := 3000.0
		transport := 1500.0
		total := rent + groceries + transport
		burden := (total / income) * 100

		result = append(result, burdenRow{
			Locality: locality,
			AvgRent:  rent,
			Total:    total,
			Burden:   burden,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"income": income, "localities": result})
}
