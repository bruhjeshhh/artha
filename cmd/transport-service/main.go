package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

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

	http.HandleFunc("/route", handleRoute)
	http.HandleFunc("/isochrone", handleIsochrone)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("transport-service listening on :8084")
	log.Fatal(http.ListenAndServe(":8084", nil))
}

func initTables(c *sql.DB) {
	_, err := c.Exec(`
		CREATE TABLE IF NOT EXISTS transport_routes (
			id SERIAL PRIMARY KEY,
			from_locality VARCHAR(100),
			to_locality VARCHAR(100),
			distance DECIMAL(10,2),
			fare DECIMAL(10,2)
		);
	`)
	if err != nil {
		log.Fatal("create table:", err)
	}
}

func seedMockData(c *sql.DB) {
	var count int
	c.QueryRow("SELECT COUNT(*) FROM transport_routes").Scan(&count)
	if count > 0 {
		return
	}

	localities := []string{"Ashta Central", "Railway Colony", "Industrial Area", "Market Ward", "Gandhi Nagar", "Nehru Colony"}

	for _, from := range localities {
		for _, to := range localities {
			if from != to {
				distance := rand.Float64()*8 + 2
				fare := distance * 8
				c.Exec(`INSERT INTO transport_routes (from_locality, to_locality, distance, fare) 
					VALUES ($1, $2, $3, $4)`, from, to, distance, fare)
			}
		}
	}
}

func handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		http.Error(w, "from and to required", http.StatusBadRequest)
		return
	}

	var route models.TransportRoute
	err := conn.QueryRow(`
		SELECT id, from_locality, to_locality, distance, fare
		FROM transport_routes 
		WHERE from_locality LIKE $1 AND to_locality LIKE $2 
		LIMIT 1
	`, "%"+from+"%", "%"+to+"%").Scan(&route.ID, &route.FromLocality, &route.ToLocality, &route.Distance, &route.Fare)

	if err == sql.ErrNoRows {
		// Return a placeholder so CLI can use commute distance
		json.NewEncoder(w).Encode(map[string]interface{}{
			"found": false,
			"from":  from,
			"to":    to,
		})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dailyCost := route.Fare * 2
	monthlyCost := dailyCost * 26
	json.NewEncoder(w).Encode(map[string]interface{}{
		"found":        true,
		"route":        route,
		"daily_cost":   dailyCost,
		"monthly_cost": monthlyCost,
	})
}

func handleIsochrone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	from := r.URL.Query().Get("from")
	if from == "" {
		http.Error(w, "from required", http.StatusBadRequest)
		return
	}

	rows, err := conn.Query(`
		SELECT to_locality, distance, fare
		FROM transport_routes
		WHERE from_locality LIKE $1
		ORDER BY distance
	`, "%"+from+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type zone struct {
		ToLocality string  `json:"to_locality"`
		Distance   float64 `json:"distance_km"`
		Fare       float64 `json:"fare"`
		TravelMin  int     `json:"travel_time_min"`
		Zone       string  `json:"time_zone"`
	}

	var result []zone
	for rows.Next() {
		var toLoc string
		var dist, fare float64
		rows.Scan(&toLoc, &dist, &fare)

		travelMin := int(dist / 25 * 60)
		zoneStr := "15 min"
		if travelMin > 30 {
			zoneStr = "45+ min"
		} else if travelMin > 15 {
			zoneStr = "30 min"
		}

		result = append(result, zone{
			ToLocality: toLoc,
			Distance:   dist,
			Fare:       fare,
			TravelMin:  travelMin,
			Zone:       zoneStr,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"from": from, "destinations": result})
}
