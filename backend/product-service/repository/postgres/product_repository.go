package postgres

import (
    "context"
    "database/sql"
    "errors"
    "time"

    "github.com/lib/pq"
    "go.uber.org/zap"
    
    "github.com/louai60/e-commerce_project/backend/product-service/models"
)

type ProductRepository struct {
    db     *sql.DB
    logger *zap.Logger
}

func NewProductRepository(db *sql.DB, logger *zap.Logger) *ProductRepository {
    return &ProductRepository{
        db:     db,
        logger: logger,
    }
}

func (r *ProductRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
    query := `
        SELECT id, name, description, price, image_url, category_id, stock, created_at, updated_at
        FROM products
        WHERE id = $1 AND deleted_at IS NULL
    `
    
    product := &models.Product{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
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

    if err == sql.ErrNoRows {
        return nil, errors.New("product not found")
    }
    if err != nil {
        return nil, err
    }

    return product, nil
}

func (r *ProductRepository) ListProducts(ctx context.Context, page, limit int32) ([]*models.Product, int64, error) {
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

    var products []*models.Product
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

func (r *ProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
    query := `
        INSERT INTO products (id, name, description, price, image_url, category_id, stock, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
    `

    now := time.Now()
    product.CreatedAt = now
    product.UpdatedAt = now

    _, err := r.db.ExecContext(ctx, query,
        product.ID,
        product.Name,
        product.Description,
        product.Price,
        product.ImageURL,
        product.CategoryID,
        product.Stock,
        now,
    )

    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok {
            switch pqErr.Code.Name() {
            case "unique_violation":
                return errors.New("product already exists")
            }
        }
        return err
    }

    return nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
    query := `
        UPDATE products
        SET name = $1,
            description = $2,
            price = $3,
            image_url = $4,
            category_id = $5,
            stock = $6,
            updated_at = $7
        WHERE id = $8 AND deleted_at IS NULL
    `

    now := time.Now()
    result, err := r.db.ExecContext(ctx, query,
        product.Name,
        product.Description,
        product.Price,
        product.ImageURL,
        product.CategoryID,
        product.Stock,
        now,
        product.ID,
    )

    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("product not found")
    }

    product.UpdatedAt = now
    return nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
    query := `
        UPDATE products
        SET deleted_at = $1
        WHERE id = $2 AND deleted_at IS NULL
    `

    result, err := r.db.ExecContext(ctx, query, time.Now(), id)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("product not found")
    }

    return nil
}