package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string
	connStr := "postgres://postgres:root@localhost:5432/nexcart_product?sslmode=disable"
	
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
	log.Println("Successfully connected to database")

	// Get the most recent product
	var productID string
	err = db.QueryRowContext(ctx, "SELECT id FROM products ORDER BY created_at DESC LIMIT 1").Scan(&productID)
	if err != nil {
		log.Fatalf("Failed to get most recent product: %v", err)
	}
	log.Printf("Most recent product ID: %s", productID)

	// Check if this product has images
	var imageCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_images WHERE product_id = $1", productID).Scan(&imageCount)
	if err != nil {
		log.Fatalf("Failed to count images: %v", err)
	}
	log.Printf("Number of images for product %s: %d", productID, imageCount)

	// If there are images, get their details
	if imageCount > 0 {
		rows, err := db.QueryContext(ctx, `
			SELECT id, url, alt_text, position, created_at, updated_at
			FROM product_images
			WHERE product_id = $1
			ORDER BY position
		`, productID)
		if err != nil {
			log.Fatalf("Failed to query images: %v", err)
		}
		defer rows.Close()

		log.Println("Images for this product:")
		log.Println("---------------------------")
		for rows.Next() {
			var id, url, altText string
			var position int
			var createdAt, updatedAt time.Time
			if err := rows.Scan(&id, &url, &altText, &position, &createdAt, &updatedAt); err != nil {
				log.Fatalf("Failed to scan image row: %v", err)
			}
			log.Printf("ID: %s, URL: %s, Alt Text: %s, Position: %d", id, url, altText, position)
		}
		if err := rows.Err(); err != nil {
			log.Fatalf("Error iterating image rows: %v", err)
		}
	}
}
