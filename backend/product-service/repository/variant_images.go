package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
	"go.uber.org/zap"
)

// AddVariantImage adds a new image to a variant
func (r *PostgresProductRepository) AddVariantImage(ctx context.Context, image *models.VariantImage) error {
	now := time.Now().UTC()
	image.CreatedAt = now
	image.UpdatedAt = now

	query := `
		INSERT INTO variant_images (variant_id, url, alt_text, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(
		ctx, query,
		image.VariantID, image.URL, image.AltText, image.Position, now, now,
	).Scan(&image.ID)

	if err != nil {
		r.logger.Error("failed to add variant image", zap.Error(err))
		return fmt.Errorf("failed to add variant image: %w", err)
	}

	return nil
}

// GetVariantImages gets all images for a variant
func (r *PostgresProductRepository) GetVariantImages(ctx context.Context, variantID string) ([]models.VariantImage, error) {
	query := `
		SELECT id, variant_id, url, alt_text, position, created_at, updated_at
		FROM variant_images
		WHERE variant_id = $1
		ORDER BY position ASC`

	rows, err := r.db.QueryContext(ctx, query, variantID)
	if err != nil {
		r.logger.Error("failed to get variant images", zap.Error(err))
		return nil, fmt.Errorf("failed to get variant images: %w", err)
	}
	defer rows.Close()

	var images []models.VariantImage
	for rows.Next() {
		var img models.VariantImage
		err := rows.Scan(
			&img.ID, &img.VariantID, &img.URL, &img.AltText, &img.Position,
			&img.CreatedAt, &img.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan variant image", zap.Error(err))
			return nil, fmt.Errorf("failed to scan variant image: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating variant images rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating variant images rows: %w", err)
	}

	return images, nil
}

// UpdateVariantImage updates an existing variant image
func (r *PostgresProductRepository) UpdateVariantImage(ctx context.Context, image *models.VariantImage) error {
	now := time.Now().UTC()
	image.UpdatedAt = now

	query := `
		UPDATE variant_images
		SET url = $1, alt_text = $2, position = $3, updated_at = $4
		WHERE id = $5`

	result, err := r.db.ExecContext(
		ctx, query,
		image.URL, image.AltText, image.Position, now, image.ID,
	)
	if err != nil {
		r.logger.Error("failed to update variant image", zap.Error(err))
		return fmt.Errorf("failed to update variant image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variant image not found")
	}

	return nil
}

// DeleteVariantImage deletes a variant image
func (r *PostgresProductRepository) DeleteVariantImage(ctx context.Context, id string) error {
	query := `DELETE FROM variant_images WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete variant image", zap.Error(err))
		return fmt.Errorf("failed to delete variant image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variant image not found")
	}

	return nil
}
