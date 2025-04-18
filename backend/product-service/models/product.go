package models

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrVariantNotFound      = errors.New("variant not found")
	ErrVariantAlreadyExists = errors.New("variant already exists")
)

type Brand struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Added for soft delete
}

type Category struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	ParentID    *string    `json:"parent_id" db:"parent_id"`
	ParentName  string     `json:"parent_name,omitempty" db:"-"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Added for soft delete
}

type ProductImage struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id"`
	URL       string    `json:"url" db:"url"`
	AltText   string    `json:"alt_text" db:"alt_text"`
	Position  int       `json:"position" db:"position"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Attribute defines the structure for product attributes like 'Color', 'Size'.
type Attribute struct {
	ID        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"` // e.g., 'Color', 'Size'
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// VariantAttributeValue holds the specific value for an attribute on a variant.
type VariantAttributeValue struct {
	Name  string `json:"name"`  // e.g., 'Color'
	Value string `json:"value"` // e.g., 'Red'
}

// ProductVariant represents a specific version of a product (e.g., Red T-Shirt, Size L).
type ProductVariant struct {
	ID            string     `json:"id" db:"id"`
	ProductID     string     `json:"product_id" db:"product_id"`
	SKU           string     `json:"sku" db:"sku"`
	Title         *string    `json:"title,omitempty" db:"title"` // Optional: "Red - Large"
	Price         float64    `json:"price" db:"price"`
	DiscountPrice *float64   `json:"discount_price,omitempty" db:"discount_price"`
	InventoryQty  int        `json:"inventory_qty" db:"inventory_qty"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Related entities (not stored directly in product_variants table)
	Attributes []VariantAttributeValue `json:"attributes,omitempty" db:"-"` // Populated via join
}

// Product represents the core product entity.
type Product struct {
	ID               string `json:"id" db:"id"`
	Title            string `json:"title" db:"title"`
	Slug             string `json:"slug" db:"slug"`
	Description      string `json:"description" db:"description"`
	ShortDescription string `json:"short_description" db:"short_description"`
	// Legacy fields needed for repository compatibility
	Price            float64    `json:"price" db:"price"`
	DiscountPrice    *float64   `json:"discount_price,omitempty" db:"discount_price"`
	SKU              string     `json:"sku" db:"sku"`
	InventoryQty     int        `json:"inventory_qty" db:"inventory_qty"`
	InventoryStatus  string     `json:"inventory_status" db:"inventory_status"` // 'in_stock', 'out_of_stock', etc.
	Weight           *float64   `json:"weight" db:"weight"`                     // Weight might stay at product level if consistent across variants
	IsPublished      bool       `json:"is_published" db:"is_published"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Added via migration 000003
	BrandID          *string    `json:"brand_id" db:"brand_id"`
	DefaultVariantID *string    `json:"default_variant_id,omitempty" db:"default_variant_id"` // Added via migration 000005

	// Related entities (populated separately)
	Brand              *Brand                 `json:"brand,omitempty" db:"-"`
	Categories         []Category             `json:"categories,omitempty" db:"-"`
	Images             []ProductImage         `json:"images,omitempty" db:"-"`   // Base product images
	Variants           []ProductVariant       `json:"variants,omitempty" db:"-"` // Product variants
	Tags               []ProductTag           `json:"tags,omitempty" db:"-"`
	Attributes         []ProductAttribute     `json:"attributes,omitempty" db:"-"`
	Specifications     []ProductSpecification `json:"specifications,omitempty" db:"-"`
	SEO                *ProductSEO            `json:"seo,omitempty" db:"-"`
	Shipping           *ProductShipping       `json:"shipping,omitempty" db:"-"`
	Discount           *ProductDiscount       `json:"discount,omitempty" db:"-"`
	InventoryLocations []InventoryLocation    `json:"inventory_locations,omitempty" db:"-"`
}

// ProductTag represents a tag associated with a product
type ProductTag struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id"`
	Tag       string    `json:"tag" db:"tag"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProductAttribute represents a product-level attribute
type ProductAttribute struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id"`
	Name      string    `json:"name" db:"name"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProductSpecification represents a technical specification for a product
type ProductSpecification struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id"`
	Name      string    `json:"name" db:"name"`
	Value     string    `json:"value" db:"value"`
	Unit      string    `json:"unit,omitempty" db:"unit"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProductSEO represents SEO metadata for a product
type ProductSEO struct {
	ID              string    `json:"id" db:"id"`
	ProductID       string    `json:"product_id" db:"product_id"`
	MetaTitle       string    `json:"meta_title" db:"meta_title"`
	MetaDescription string    `json:"meta_description" db:"meta_description"`
	Keywords        []string  `json:"keywords" db:"keywords"`
	Tags            []string  `json:"tags" db:"tags"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ProductShipping represents shipping information for a product
type ProductShipping struct {
	ID               string    `json:"id" db:"id"`
	ProductID        string    `json:"product_id" db:"product_id"`
	FreeShipping     bool      `json:"free_shipping" db:"free_shipping"`
	EstimatedDays    int       `json:"estimated_days" db:"estimated_days"`
	ExpressAvailable bool      `json:"express_available" db:"express_available"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ProductDiscount represents a discount for a product
type ProductDiscount struct {
	ID        string     `json:"id" db:"id"`
	ProductID string     `json:"product_id" db:"product_id"`
	Type      string     `json:"type" db:"discount_type"`
	Value     float64    `json:"value" db:"value"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// InventoryLocation represents a warehouse location for a product's inventory
type InventoryLocation struct {
	ID           string    `json:"id" db:"id"`
	ProductID    string    `json:"product_id" db:"product_id"`
	WarehouseID  string    `json:"warehouse_id" db:"warehouse_id"`
	AvailableQty int       `json:"available_qty" db:"available_qty"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ProductFilters struct {
	Category  string   `json:"category"` // TODO: Update filters based on variants/attributes in Phase 5
	PriceMin  float64  `json:"price_min"`
	PriceMax  float64  `json:"price_max"`
	Tags      []string `json:"tags"`
	SortBy    string   `json:"sort_by"`
	SortOrder string   `json:"sort_order"`
	Page      int      `json:"page"`
	PageSize  int      `json:"page_size"`
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
