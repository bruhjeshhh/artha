package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
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

	initTables(conn)
	seedMockData(conn)

	http.HandleFunc("/data", handleData)
	http.HandleFunc("/summary", handleSummary)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("inflation-service listening on :8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}

func initTables(c *sql.DB) {
	_, err := c.Exec(`
		CREATE TABLE IF NOT EXISTS inflation_data (
			id SERIAL PRIMARY KEY,
			month VARCHAR(20),
			rate DECIMAL(5,2),
			category VARCHAR(50)
		);
	`)
	if err != nil {
		log.Fatal("create table:", err)
	}
}

func seedMockData(c *sql.DB) {
	var count int
	c.QueryRow("SELECT COUNT(*) FROM inflation_data").Scan(&count)
	if count > 0 {
		return
	}

	months := []string{"Jan 2025", "Dec 2024", "Nov 2024", "Oct 2024", "Sep 2024", "Aug 2024"}
	categories := []string{"Food", "Housing", "Transport", "Overall"}

	for _, month := range months {
		for _, category := range categories {
			rate := 5.5 + rand.Float64()*2.5
			c.Exec(`INSERT INTO inflation_data (month, rate, category) VALUES ($1, $2, $3)`,
				month, rate, category)
		}
	}
}

func handleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	rows, err := conn.Query(`
		SELECT month, category, rate 
		FROM inflation_data 
		ORDER BY month DESC, category
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type row struct {
		Month    string  `json:"month"`
		Category string  `json:"category"`
		Rate     float64 `json:"rate"`
	}

	var list []row
	for rows.Next() {
		var r row
		rows.Scan(&r.Month, &r.Category, &r.Rate)
		list = append(list, r)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"data": list})
}

func handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var avgRate float64
	err := conn.QueryRow("SELECT AVG(rate) FROM inflation_data WHERE category = 'Overall'").Scan(&avgRate)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"average_overall_inflation": avgRate,
		"trend":                     "Inflation has been relatively stable over the past 6 months",
	})
}
