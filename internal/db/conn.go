package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

// DefaultConnStr is the default PostgreSQL connection string for local dev.
const DefaultConnStr = "host=localhost port=5433 user=postgres password=postgres dbname=rentanalyzer sslmode=disable"

// ConnStr returns the DB connection string from env or default.
func ConnStr() string {
	if s := os.Getenv("DB_URL"); s != "" {
		return s
	}
	return DefaultConnStr
}

// Open opens a PostgreSQL connection using ConnStr().
func Open() (*sql.DB, error) {
	return sql.Open("postgres", ConnStr())
}
