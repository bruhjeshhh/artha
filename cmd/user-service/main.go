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

	http.HandleFunc("/profile", handleProfile)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	log.Println("user-service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func initTables(c *sql.DB) {
	_, err := c.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY DEFAULT 1,
			name VARCHAR(100),
			income DECIMAL(12,2),
			family_size INT,
			preferred_locale VARCHAR(100),
			commute_distance DECIMAL(10,2)
		);
	`)
	if err != nil {
		log.Fatal("create table:", err)
	}
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		var u models.UserProfile
		err := conn.QueryRow(`
			SELECT id, name, income, family_size, preferred_locale, commute_distance
			FROM users WHERE id = 1
		`).Scan(&u.ID, &u.Name, &u.Income, &u.FamilySize, &u.PreferredLocale, &u.CommuteDistance)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "no profile"})
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(u)
		return

	case http.MethodPost:
		var u models.UserProfile
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := conn.Exec(`
			INSERT INTO users (id, name, income, family_size, preferred_locale, commute_distance)
			VALUES (1, $1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				name = $1, income = $2, family_size = $3, preferred_locale = $4, commute_distance = $5
		`, u.Name, u.Income, u.FamilySize, u.PreferredLocale, u.CommuteDistance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u.ID = 1
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
