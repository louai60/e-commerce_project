package main

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"
// 	"time"

// 	"github.com/joho/godotenv"
// 	_ "github.com/lib/pq"
// 	"github.com/louai60/e-commerce_project/backend/common/logger"
// 	"github.com/louai60/e-commerce_project/backend/product-service/config"
// 	"go.uber.org/zap"
// )

// func main() {
// 	// Load .env file
// 	if err := godotenv.Load(); err != nil {
// 		// Only log error if .env file exists but couldn't be loaded
// 		if !os.IsNotExist(err) {
// 			log.Fatalf("Error loading .env file: %v", err)
// 		}
// 	}

// 	// Initialize logger
// 	zapLogger := logger.GetLogger()
// 	defer zapLogger.Sync()

// 	// Load configuration
// 	cfg, err := config.LoadConfig(zapLogger)
// 	if err != nil {
// 		zapLogger.Fatal("Failed to load configuration", zap.Error(err))
// 	}

// 	// Initialize database connection
// 	db, err := sql.Open("postgres", cfg.GetDSN())
// 	if err != nil {
// 		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
// 	}
// 	defer db.Close()

// 	// Set connection pool parameters
// 	db.SetMaxOpenConns(25)
// 	db.SetMaxIdleConns(5)
// 	db.SetConnMaxLifetime(5 * time.Minute)

// 	// Test database connection
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	if err := db.PingContext(ctx); err != nil {
// 		zapLogger.Fatal("Failed to ping database", zap.Error(err))
// 	}

// 	// First, create a test product
// 	productID, err := createTestProduct(ctx, db, zapLogger)
// 	if err != nil {
// 		zapLogger.Fatal("Failed to create test product", zap.Error(err))
// 	}
// 	zapLogger.Info("Created test product", zap.String("id", productID))

// 	// Now, insert a test image for this product
// 	err = insertTestImage(ctx, db, productID, zapLogger)
// 	if err != nil {
// 		zapLogger.Fatal("Failed to insert test image", zap.Error(err))
// 	}
// 	zapLogger.Info("Successfully inserted test image")

// 	// Verify the image was inserted
// 	images, err := getProductImages(ctx, db, productID, zapLogger)
// 	if err != nil {
// 		zapLogger.Fatal("Failed to get product images", zap.Error(err))
// 	}

// 	if len(images) == 0 {
// 		zapLogger.Error("No images found for product", zap.String("product_id", productID))
// 	} else {
// 		zapLogger.Info("Found images for product",
// 			zap.String("product_id", productID),
// 			zap.Int("count", len(images)))

// 		for i, img := range images {
// 			zapLogger.Info(fmt.Sprintf("Image %d", i+1),
// 				zap.String("id", img.ID),
// 				zap.String("url", img.URL),
// 				zap.String("alt_text", img.AltText))
// 		}
// 	}
// }

// func createTestProduct(ctx context.Context, db *sql.DB, logger *zap.Logger) (string, error) {
// 	now := time.Now()

// 	query := `
// 		INSERT INTO products (
// 			title, slug, description, short_description, price,
// 			sku, inventory_qty, inventory_status, is_published, created_at, updated_at
// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
// 		RETURNING id`

// 	var productID string
// 	err := db.QueryRowContext(
// 		ctx, query,
// 		"Test Product for Image Test",
// 		"test-product-image-test",
// 		"This is a test product for image testing",
// 		"Test product for image testing",
// 		99.99,
// 		fmt.Sprintf("TEST-SKU-%d", time.Now().Unix()),
// 		100,
// 		"in_stock",
// 		true,
// 		now,
// 		now,
// 	).Scan(&productID)

// 	if err != nil {
// 		logger.Error("Failed to create test product", zap.Error(err))
// 		return "", err
// 	}

// 	return productID, nil
// }

// func insertTestImage(ctx context.Context, db *sql.DB, productID string, logger *zap.Logger) error {
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
// 		logger.Error("Failed to insert test image", zap.Error(err))
// 		return err
// 	}

// 	logger.Info("Inserted test image", zap.String("id", imageID))
// 	return nil
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

// func getProductImages(ctx context.Context, db *sql.DB, productID string, logger *zap.Logger) ([]ProductImage, error) {
// 	query := `
// 		SELECT id, product_id, url, alt_text, position, created_at, updated_at
// 		FROM product_images
// 		WHERE product_id = $1
// 		ORDER BY position`

// 	rows, err := db.QueryContext(ctx, query, productID)
// 	if err != nil {
// 		logger.Error("Failed to query product images", zap.Error(err))
// 		return nil, err
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
// 			logger.Error("Failed to scan image row", zap.Error(err))
// 			return nil, err
// 		}
// 		images = append(images, img)
// 	}

// 	if err = rows.Err(); err != nil {
// 		logger.Error("Error iterating image rows", zap.Error(err))
// 		return nil, err
// 	}

// 	return images, nil
// }
