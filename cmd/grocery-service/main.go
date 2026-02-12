package main

import (
	"database/sql"
	"encoding/json"
	"log"
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

	http.HandleFunc("/items", handleItems)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("grocery-service listening on :8083")
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func initTables(c *sql.DB) {
	_, err := c.Exec(`
		CREATE TABLE IF NOT EXISTS groceries (
			id SERIAL PRIMARY KEY,
			item VARCHAR(100),
			price DECIMAL(10,2),
			source VARCHAR(50)
		);
	`)
	if err != nil {
		log.Fatal("create table:", err)
	}
}

func seedMockData(c *sql.DB) {
	var count int
	c.QueryRow("SELECT COUNT(*) FROM groceries").Scan(&count)
	if count > 0 {
		return
	}

	items := []struct {
		item   string
		price  float64
		source string
	}{
		{"Rice (1kg)", 45.0, "BigBasket"},
		{"Wheat Flour (1kg)", 40.0, "Blinkit"},
		{"Cooking Oil (1L)", 150.0, "BigBasket"},
		{"Milk (1L)", 55.0, "Blinkit"},
		{"Vegetables (weekly)", 300.0, "BigBasket"},
		{"Lentils (1kg)", 80.0, "Blinkit"},
		{"Sugar (1kg)", 42.0, "BigBasket"},
		{"Tea/Coffee", 120.0, "Blinkit"},
	}

	for _, it := range items {
		c.Exec(`INSERT INTO groceries (item, price, source) VALUES ($1, $2, $3)`,
			it.item, it.price, it.source)
	}
}

func handleItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	rows, err := conn.Query("SELECT item, price, source FROM groceries ORDER BY price DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []models.GroceryItem
	var total float64
	for rows.Next() {
		var g models.GroceryItem
		err := rows.Scan(&g.Item, &g.Price, &g.Source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, g)
		total += g.Price
	}

	monthlyEstimate := total * 4.3
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":            list,
		"total_basket":     total,
		"monthly_estimate": monthlyEstimate,
	})
}
