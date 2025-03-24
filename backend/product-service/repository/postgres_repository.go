package repository

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `
        INSERT INTO products (id, name, description, price, image_url, category_id, stock, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.CategoryID,
		product.Stock,
		product.CreatedAt,
		product.UpdatedAt,
	)

	return err
}

func (r *PostgresRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	query := `
        SELECT id, name, description, price, image_url, category_id, stock, created_at, updated_at 
        FROM products WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	product := &models.Product{}
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.ImageURL,
		&product.CategoryID,
		&product.Stock,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return product, nil
}

func (r *PostgresRepository) ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get products
	query := `
        SELECT id, name, description, price, image_url, category_id, stock, created_at, updated_at
        FROM products
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := []*models.Product{}
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.ImageURL,
			&product.CategoryID,
			&product.Stock,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	return products, total, nil
}

func (r *PostgresRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	query := `
        UPDATE products SET
            name = $2,
            description = $3,
            price = $4,
            image_url = $5,
            category_id = $6,
            stock = $7,
            updated_at = $8
        WHERE id = $1
    `
	_, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.CategoryID,
		product.Stock,
		product.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ping is already implemented in repository/repository.go
