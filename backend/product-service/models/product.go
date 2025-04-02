package models

import (
	"fmt"
	"sort"
	"strings"
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

type ProductFilters struct {
	Category   string   `json:"category"`
	PriceMin   float64  `json:"price_min"`
	PriceMax   float64  `json:"price_max"`
	Tags       []string `json:"tags"`
	SortBy     string   `json:"sort_by"`
	SortOrder  string   `json:"sort_order"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

func (f *ProductFilters) ToCacheKey() string {
	components := []string{
		fmt.Sprintf("cat:%s", f.Category),
		fmt.Sprintf("price:%.2f-%.2f", f.PriceMin, f.PriceMax),
	}

	if len(f.Tags) > 0 {
		sort.Strings(f.Tags)
		components = append(components, fmt.Sprintf("tags:%s", strings.Join(f.Tags, ",")))
	}

	components = append(components,
		fmt.Sprintf("sort:%s:%s", f.SortBy, f.SortOrder),
		fmt.Sprintf("page:%d:%d", f.Page, f.PageSize),
	)

	return strings.Join(components, "|")
}
