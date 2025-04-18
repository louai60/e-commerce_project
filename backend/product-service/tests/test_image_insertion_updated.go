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

// 	// First, create a test product
// 	productID, err := createTestProduct(ctx, db)
// 	if err != nil {
// 		log.Fatalf("Failed to create test product: %v", err)
// 	}
// 	log.Printf("Created test product with ID: %s", productID)

// 	// Create a variant for the product
// 	variantID, err := createTestVariant(ctx, db, productID)
// 	if err != nil {
// 		log.Fatalf("Failed to create test variant: %v", err)
// 	}
// 	log.Printf("Created test variant with ID: %s", variantID)

// 	// Now, insert a test image for this product
// 	imageID, err := insertTestImage(ctx, db, productID)
// 	if err != nil {
// 		log.Fatalf("Failed to insert test image: %v", err)
// 	}
// 	log.Printf("Successfully inserted test image with ID: %s", imageID)

// 	// Verify the image was inserted
// 	images, err := getProductImages(ctx, db, productID)
// 	if err != nil {
// 		log.Fatalf("Failed to get product images: %v", err)
// 	}

// 	if len(images) == 0 {
// 		log.Printf("No images found for product %s", productID)
// 	} else {
// 		log.Printf("Found %d images for product %s", len(images), productID)

// 		for i, img := range images {
// 			log.Printf("Image %d: ID=%s, URL=%s, AltText=%s",
// 				i+1, img.ID, img.URL, img.AltText)
// 		}
// 	}
// }

// func createTestProduct(ctx context.Context, db *sql.DB) (string, error) {
// 	now := time.Now()

// 	query := `
// 		INSERT INTO products (
// 			title, slug, description, short_description,
// 			is_published, inventory_status, created_at, updated_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
// 		RETURNING id`

// 	var productID string
// 	err := db.QueryRowContext(
// 		ctx, query,
// 		"Test Product for Image Test",
// 		"test-product-image-test-"+fmt.Sprintf("%d", time.Now().Unix()),
// 		"This is a test product for image testing",
// 		"Test product for image testing",
// 		true,
// 		"in_stock",
// 		now,
// 		now,
// 	).Scan(&productID)

// 	if err != nil {
// 		return "", fmt.Errorf("failed to create test product: %w", err)
// 	}

// 	return productID, nil
// }

// func createTestVariant(ctx context.Context, db *sql.DB, productID string) (string, error) {
// 	now := time.Now()

// 	query := `
// 		INSERT INTO product_variants (
// 			product_id, title, sku, price, inventory_qty,
// 			created_at, updated_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		RETURNING id`

// 	var variantID string
// 	err := db.QueryRowContext(
// 		ctx, query,
// 		productID,
// 		"Test Variant",
// 		fmt.Sprintf("TEST-SKU-%d", time.Now().Unix()),
// 		99.99,
// 		100,
// 		now,
// 		now,
// 	).Scan(&variantID)

// 	if err != nil {
// 		return "", fmt.Errorf("failed to create test variant: %w", err)
// 	}

// 	// Update the product's default_variant_id
// 	updateQuery := `
// 		UPDATE products
// 		SET default_variant_id = $1
// 		WHERE id = $2`

// 	_, err = db.ExecContext(ctx, updateQuery, variantID, productID)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to update product with default variant: %w", err)
// 	}

// 	return variantID, nil
// }

// func insertTestImage(ctx context.Context, db *sql.DB, productID string) (string, error) {
// 	now := time.Now()

// 	query := `
// 		INSERT INTO product_images (
// 			product_id, url, alt_text, position, created_at, updated_at
// 		) VALUES ($1, $2, $3, $4, $5, $6)
// 		RETURNING id`

// 	var imageID string
// 	err := db.QueryRowContext(
// 		ctx, query,
// 		productID,
// 		"https://example.com/test-image.jpg",
// 		"Test Image",
// 		0,
// 		now,
// 		now,
// 	).Scan(&imageID)

// 	if err != nil {
// 		return "", fmt.Errorf("failed to insert test image: %w", err)
// 	}

// 	return imageID, nil
// }

// type ProductImage struct {
// 	ID        string
// 	ProductID string
// 	URL       string
// 	AltText   string
// 	Position  int
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// }

// func getProductImages(ctx context.Context, db *sql.DB, productID string) ([]ProductImage, error) {
// 	query := `
// 		SELECT id, product_id, url, alt_text, position, created_at, updated_at
// 		FROM product_images
// 		WHERE product_id = $1
// 		ORDER BY position`

// 	rows, err := db.QueryContext(ctx, query, productID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query product images: %w", err)
// 	}
// 	defer rows.Close()

// 	var images []ProductImage
// 	for rows.Next() {
// 		var img ProductImage
// 		err := rows.Scan(
// 			&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.Position,
// 			&img.CreatedAt, &img.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to scan image row: %w", err)
// 		}
// 		images = append(images, img)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("error iterating image rows: %w", err)
// 	}

// 	return images, nil
// }
