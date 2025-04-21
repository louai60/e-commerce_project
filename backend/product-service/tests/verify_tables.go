package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Verifying database tables...")

	// Database connection string
	connStr := "postgres://postgres:root@localhost:5432/nexcart_product?sslmode=disable"
	fmt.Println("Using connection string:", connStr)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to database")

	// Write output to a file as well
	logFile, err := os.Create("table_verification.log")
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	logFile.WriteString("Database Table Verification\n")
	logFile.WriteString("=========================\n\n")

	// Tables to check
	tables := []string{
		"products",
		"product_variants",
		"product_images",
		"variant_images",
		"product_categories",
		"product_tags",
		"product_specifications",
		"product_attributes",
		"product_seo",
		"product_shipping",
		"product_discounts",
		"product_inventory_locations",
		"schema_migrations",
	}

	fmt.Println("\nTable Status:")
	fmt.Println("---------------------------")
	fmt.Printf("%-30s %-10s\n", "Table Name", "Exists")
	fmt.Println("---------------------------")

	for _, table := range tables {
		var exists bool
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = $1
			)
		`, table).Scan(&exists)

		if err != nil {
			log.Fatalf("Failed to check if %s table exists: %v", table, err)
		}

		fmt.Printf("%-30s %-10t\n", table, exists)
	}

	// Check migration status
	rows, err := db.QueryContext(ctx, `SELECT version, dirty FROM schema_migrations`)
	if err != nil {
		log.Fatalf("Failed to query schema_migrations: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nMigration Status:")
	fmt.Println("---------------------------")
	fmt.Printf("%-10s %-10s\n", "Version", "Dirty")
	fmt.Println("---------------------------")

	for rows.Next() {
		var version int
		var dirty bool
		if err := rows.Scan(&version, &dirty); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("%-10d %-10t\n", version, dirty)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}
}
