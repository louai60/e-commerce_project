package repository

import (
	"context"
	"fmt"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

// GetProductImages retrieves all images for a product
func (r *PostgresProductRepository) GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error) {
	query := `
		SELECT id, product_id, url, alt_text, position, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY position ASC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product images", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}
	defer rows.Close()

	var images []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		err := rows.Scan(
			&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.Position,
			&img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product image", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product images rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product images rows: %w", err)
	}

	return images, nil
}

// GetProductImages retrieves all images for a product
func (r *PostgresRepository) GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error) {
	query := `
		SELECT id, product_id, url, alt_text, position, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY position ASC`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		r.logger.Error("failed to get product images", zap.Error(err), zap.String("product_id", productID))
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}
	defer rows.Close()

	var images []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		err := rows.Scan(
			&img.ID, &img.ProductID, &img.URL, &img.AltText, &img.Position,
			&img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan product image", zap.Error(err))
			return nil, fmt.Errorf("failed to scan product image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating product images rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating product images rows: %w", err)
	}

	return images, nil
}

// getProductImages fetches all images associated with a product
// func (r *PostgresRepository) getProductImages(ctx context.Context, product *models.Product) error {
// 	images, err := r.GetProductImages(ctx, product.ID)
// 	if err != nil {
// 		return err
// 	}
// 	product.Images = images
// 	return nil
// }
