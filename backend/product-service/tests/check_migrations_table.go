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

// 	// Check if schema_migrations table exists
// 	var exists bool
// 	err = db.QueryRowContext(ctx, `
// 		SELECT EXISTS (
// 			SELECT FROM information_schema.tables 
// 			WHERE table_name = 'schema_migrations'
// 		)
// 	`).Scan(&exists)
	
// 	if err != nil {
// 		log.Fatalf("Failed to check if schema_migrations table exists: %v", err)
// 	}
	
// 	if !exists {
// 		log.Println("schema_migrations table does not exist")
// 		return
// 	}
	
// 	// Get schema_migrations table content
// 	rows, err := db.QueryContext(ctx, `SELECT version, dirty FROM schema_migrations`)
// 	if err != nil {
// 		log.Fatalf("Failed to query schema_migrations: %v", err)
// 	}
// 	defer rows.Close()

// 	log.Println("Schema Migrations:")
// 	log.Println("---------------------------")
// 	log.Printf("%-10s %-10s", "Version", "Dirty")
// 	log.Println("---------------------------")

// 	for rows.Next() {
// 		var version int
// 		var dirty bool
// 		if err := rows.Scan(&version, &dirty); err != nil {
// 			log.Fatalf("Failed to scan row: %v", err)
// 		}
// 		log.Printf("%-10d %-10t", version, dirty)
// 	}

// 	if err := rows.Err(); err != nil {
// 		log.Fatalf("Error iterating rows: %v", err)
// 	}
	
// 	// Check if variant_images table exists
// 	err = db.QueryRowContext(ctx, `
// 		SELECT EXISTS (
// 			SELECT FROM information_schema.tables 
// 			WHERE table_name = 'variant_images'
// 		)
// 	`).Scan(&exists)
	
// 	if err != nil {
// 		log.Fatalf("Failed to check if variant_images table exists: %v", err)
// 	}
	
// 	log.Printf("variant_images table exists: %t", exists)
// }
