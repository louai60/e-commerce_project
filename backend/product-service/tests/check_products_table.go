package main

// import (
// 	"context"
// 	"database/sql"
// 	"log"
// 	"time"

// 	_ "github.com/lib/pq"
// )

// func main() {
// 	// Database connection string
// 	connStr := "postgres://postgres:root@localhost:5432/nexcart_product?sslmode=disable"
	
// 	// Connect to database
// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer db.Close()

// 	// Test database connection
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	if err := db.PingContext(ctx); err != nil {
// 		log.Fatalf("Failed to ping database: %v", err)
// 	}
// 	log.Println("Successfully connected to database")

// 	// Get table columns
// 	query := `
// 		SELECT column_name, data_type, is_nullable
// 		FROM information_schema.columns
// 		WHERE table_name = 'products'
// 		ORDER BY ordinal_position;
// 	`

// 	rows, err := db.QueryContext(ctx, query)
// 	if err != nil {
// 		log.Fatalf("Failed to query table structure: %v", err)
// 	}
// 	defer rows.Close()

// 	log.Println("Products table structure:")
// 	log.Println("---------------------------")
// 	log.Printf("%-20s %-20s %-10s", "Column Name", "Data Type", "Nullable")
// 	log.Println("---------------------------")

// 	for rows.Next() {
// 		var columnName, dataType, isNullable string
// 		if err := rows.Scan(&columnName, &dataType, &isNullable); err != nil {
// 			log.Fatalf("Failed to scan row: %v", err)
// 		}
// 		log.Printf("%-20s %-20s %-10s", columnName, dataType, isNullable)
// 	}

// 	if err := rows.Err(); err != nil {
// 		log.Fatalf("Error iterating rows: %v", err)
// 	}
// }
