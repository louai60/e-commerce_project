package main

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"time"

// 	_ "github.com/lib/pq"
// )

// func main() {
// 	fmt.Println("Starting check for variant_images table...")
	
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
// 	fmt.Println("Successfully connected to database")

// 	// Check if variant_images table exists
// 	var exists bool
// 	err = db.QueryRowContext(ctx, `
// 		SELECT EXISTS (
// 			SELECT FROM information_schema.tables 
// 			WHERE table_name = 'variant_images'
// 		)
// 	`).Scan(&exists)
	
// 	if err != nil {
// 		log.Fatalf("Failed to check if variant_images table exists: %v", err)
// 	}
	
// 	fmt.Printf("variant_images table exists: %t\n", exists)
	
// 	// Check schema_migrations table
// 	err = db.QueryRowContext(ctx, `
// 		SELECT EXISTS (
// 			SELECT FROM information_schema.tables 
// 			WHERE table_name = 'schema_migrations'
// 		)
// 	`).Scan(&exists)
	
// 	if err != nil {
// 		log.Fatalf("Failed to check if schema_migrations table exists: %v", err)
// 	}
	
// 	fmt.Printf("schema_migrations table exists: %t\n", exists)
	
// 	if exists {
// 		// Get schema_migrations table content
// 		rows, err := db.QueryContext(ctx, `SELECT version, dirty FROM schema_migrations`)
// 		if err != nil {
// 			log.Fatalf("Failed to query schema_migrations: %v", err)
// 		}
// 		defer rows.Close()

// 		fmt.Println("Schema Migrations:")
// 		fmt.Println("---------------------------")
// 		fmt.Printf("%-10s %-10s\n", "Version", "Dirty")
// 		fmt.Println("---------------------------")

// 		for rows.Next() {
// 			var version int
// 			var dirty bool
// 			if err := rows.Scan(&version, &dirty); err != nil {
// 				log.Fatalf("Failed to scan row: %v", err)
// 			}
// 			fmt.Printf("%-10d %-10t\n", version, dirty)
// 		}

// 		if err := rows.Err(); err != nil {
// 			log.Fatalf("Error iterating rows: %v", err)
// 		}
// 	}
// }
