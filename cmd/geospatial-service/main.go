package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"rent-cost-analyzer/internal/db"
)

var conn *sql.DB

func main() {
	c, err := db.Open()
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer c.Close()
	conn = c

	// Uses rental_listings table (read-only); no tables created here

	http.HandleFunc("/heatmap", handleHeatmap)
	http.HandleFunc("/nearby", handleNearby)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("geospatial-service listening on :8086")
	log.Fatal(http.ListenAndServe(":8086", nil))
}

func handleHeatmap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	rows, err := conn.Query(`
		SELECT locality, AVG(rent) as avg_rent, COUNT(*) as count
		FROM rental_listings
		GROUP BY locality
		ORDER BY avg_rent DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type localityRow struct {
		Locality string  `json:"locality"`
		AvgRent  float64 `json:"avg_rent"`
		Count    int     `json:"count"`
	}

	var list []localityRow
	var maxRent float64
	for rows.Next() {
		var row localityRow
		rows.Scan(&row.Locality, &row.AvgRent, &row.Count)
		list = append(list, row)
		if row.AvgRent > maxRent {
			maxRent = row.AvgRent
		}
	}

	// Add intensity 0-1 for client
	type withIntensity struct {
		Locality  string  `json:"locality"`
		AvgRent   float64 `json:"avg_rent"`
		Count     int     `json:"count"`
		Intensity float64 `json:"intensity"`
	}
	var result []withIntensity
	for _, row := range list {
		val := 0.0
		if maxRent > 0 {
			val = row.AvgRent / maxRent
		}
		result = append(result, withIntensity{
			Locality: row.Locality, AvgRent: row.AvgRent, Count: row.Count,
			Intensity: val,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"localities": result})
}

func handleNearby(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	locality := r.URL.Query().Get("locality")
	if locality == "" {
		http.Error(w, "locality required", http.StatusBadRequest)
		return
	}

	rows, err := conn.Query(`
		SELECT locality, distance, lat, lon
		FROM rental_listings
		WHERE locality != $1
		ORDER BY distance
		LIMIT 10
	`, locality)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type nearbyRow struct {
		Locality  string  `json:"locality"`
		Distance  float64 `json:"distance_km"`
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
	}

	var list []nearbyRow
	for rows.Next() {
		var row nearbyRow
		rows.Scan(&row.Locality, &row.Distance, &row.Lat, &row.Lon)
		list = append(list, row)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"center": locality, "nearby": list})
}
