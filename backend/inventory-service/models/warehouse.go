package models

import (
	"time"
)

// Warehouse represents a physical location where inventory is stored
type Warehouse struct {
	ID         string    `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Code       string    `json:"code" db:"code"`
	Address    string    `json:"address" db:"address"`
	City       string    `json:"city" db:"city"`
	State      string    `json:"state" db:"state"`
	Country    string    `json:"country" db:"country"`
	PostalCode string    `json:"postal_code" db:"postal_code"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	Priority   int       `json:"priority" db:"priority"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
