package models

import (
	"time"
)

// Product represents a product in the e-commerce system
type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	CategoryID  string    `json:"category_id" db:"category_id"`
	Stock       int       `json:"stock" db:"stock"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
