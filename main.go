package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Item struct {
	ID   int
	Name string
}

func init() {
	godotenv.Load()
}

func main() {
	// Build connection string from env vars
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "db" // Use Docker Compose service name if not set
	}
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// Create table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Table creation error: %v", err)
	}

	// Insert dummy data if table is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
	if err != nil {
		log.Fatalf("Count query error: %v", err)
	}
	if count == 0 {
		_, err := db.Exec("INSERT INTO items (name) VALUES ($1), ($2), ($3)", "Apple", "Banana", "Cherry")
		if err != nil {
			log.Fatalf("Insert error: %v", err)
		}
	}

	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name FROM items")
		if err != nil {
			http.Error(w, "Database error", 500)
			return
		}
		defer rows.Close()

		var items []Item
		for rows.Next() {
			var it Item
			if err := rows.Scan(&it.ID, &it.Name); err != nil {
				http.Error(w, "Row scan error", 500)
				return
			}
			items = append(items, it)
		}

		// Print the list as plain text
		for _, it := range items {
			fmt.Fprintf(w, "ID: %d, Name: %s\n", it.ID, it.Name)
		}
	})

	port = "8080"
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
