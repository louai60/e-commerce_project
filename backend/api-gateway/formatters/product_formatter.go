package formatters

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProductResponse represents the formatted product response
type ProductResponse struct {
	ID               string                 `json:"id"`
	Title            string                 `json:"title"`
	Slug             string                 `json:"slug"`
	ShortDescription string                 `json:"short_description"`
	Description      string                 `json:"description"`
	SKU              string                 `json:"sku"`
	DefaultVariantID string                 `json:"default_variant_id,omitempty"`
	Price            *EnhancedPriceInfo     `json:"price"`
	Attributes       []AttributeInfo        `json:"attributes"`
	Variants         []EnhancedVariantInfo  `json:"variants"`
	Images           []EnhancedImageInfo    `json:"images"`
	Reviews          *EnhancedReviewInfo    `json:"reviews,omitempty"`
	Tags             []string               `json:"tags"`
	Specifications   []SpecificationInfo    `json:"specifications"`
	Brand            *BrandInfo             `json:"brand,omitempty"`
	Categories       []CategoryInfo         `json:"categories,omitempty"`
	Inventory        *EnhancedInventoryInfo `json:"inventory"`
	Metadata         *MetadataInfo          `json:"metadata"`
	SEO              *EnhancedSEOInfo       `json:"seo,omitempty"`
	Shipping         *EnhancedShippingInfo  `json:"shipping,omitempty"`
	Discounts        []DiscountInfo         `json:"discounts,omitempty"`
}

// CategoryInfo represents category information
type CategoryInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// BrandInfo represents brand information
type BrandInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// EnhancedPriceInfo represents enhanced price information
type EnhancedPriceInfo struct {
	Current           map[string]float64 `json:"current"`
	Currency          string             `json:"currency"`
	SavingsPercentage float64            `json:"savings_percentage,omitempty"`
	// Simple price value for consistency with request format
	Value float64 `json:"value,omitempty"`
}

// EnhancedInventoryInfo represents enhanced inventory information
type EnhancedInventoryInfo struct {
	Status    string                 `json:"status"`
	Available bool                   `json:"available"`
	Quantity  int                    `json:"quantity"`
	Locations []EnhancedLocationInfo `json:"locations,omitempty"`
}

// EnhancedLocationInfo represents enhanced warehouse location information
type EnhancedLocationInfo struct {
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}

// WeightInfo represents weight information
type WeightInfo struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// EnhancedImageInfo represents enhanced image information
type EnhancedImageInfo struct {
	ID          string `json:"id,omitempty"`
	URL         string `json:"url"`
	AltText     string `json:"alt_text"`
	Position    int    `json:"position"`
	ViewType    string `json:"view_type,omitempty"`
	IsThumbnail bool   `json:"is_thumbnail,omitempty"`
}

// EnhancedVariantInfo represents enhanced variant information
type EnhancedVariantInfo struct {
	ID            string              `json:"id"`
	ProductID     string              `json:"product_id"`
	SKU           string              `json:"sku"`
	Title         string              `json:"title"`
	Price         float64             `json:"price"`
	DiscountPrice float64             `json:"discount_price,omitempty"`
	InventoryQty  int                 `json:"inventory_qty"`
	Attributes    []AttributeInfo     `json:"attributes"`
	Images        []EnhancedImageInfo `json:"images"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`

	// Inherited fields from parent product
	Description      string                `json:"description,omitempty"`
	ShortDescription string                `json:"short_description,omitempty"`
	Specifications   []SpecificationInfo   `json:"specifications,omitempty"`
	Tags             []string              `json:"tags,omitempty"`
	Categories       []CategoryInfo        `json:"categories,omitempty"`
	Brand            *BrandInfo            `json:"brand,omitempty"`
	SEO              *EnhancedSEOInfo      `json:"seo,omitempty"`
	Shipping         *EnhancedShippingInfo `json:"shipping,omitempty"`
	Discounts        []DiscountInfo        `json:"discounts,omitempty"`
}

// AttributeInfo represents attribute information
type AttributeInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SpecificationInfo represents specification information
type SpecificationInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

// EnhancedReviewInfo represents enhanced review information
type EnhancedReviewInfo struct {
	Summary EnhancedReviewSummary `json:"summary"`
	Items   []EnhancedReviewItem  `json:"items"`
}

// EnhancedReviewSummary represents enhanced review summary
type EnhancedReviewSummary struct {
	AverageRating      float64        `json:"average_rating"`
	TotalReviews       int            `json:"total_reviews"`
	RatingDistribution map[string]int `json:"rating_distribution"`
}

// EnhancedReviewItem represents an enhanced review item
type EnhancedReviewItem struct {
	ID           string     `json:"id"`
	User         ReviewUser `json:"user"`
	Rating       int        `json:"rating"`
	Title        string     `json:"title"`
	Comment      string     `json:"comment"`
	Date         string     `json:"date"`
	HelpfulVotes int        `json:"helpful_votes"`
}

// ReviewUser represents a review user
type ReviewUser struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	VerifiedPurchaser bool   `json:"verified_purchaser"`
}

// EnhancedShippingInfo represents enhanced shipping information
type EnhancedShippingInfo struct {
	FreeShipping             bool   `json:"free_shipping"`
	EstimatedDays            string `json:"estimated_days"`
	ExpressShippingAvailable bool   `json:"express_shipping_available"`
	ExpressShippingDays      string `json:"express_shipping_days,omitempty"`
}

// DiscountInfo represents discount information
type DiscountInfo struct {
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	ExpiresAt string  `json:"expires_at,omitempty"`
}

// EnhancedSEOInfo represents enhanced SEO information
type EnhancedSEOInfo struct {
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	Keywords        []string `json:"keywords"`
	MetaTags        []string `json:"meta_tags"`
}

// MetadataInfo represents metadata information
type MetadataInfo struct {
	IsPublished bool  `json:"is_published"`
	CreatedAt   int64 `json:"created_at"`
	UpdatedAt   int64 `json:"updated_at"`
}

// MetaInfo represents pagination metadata
type MetaInfo struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
}

// ProductListResponse represents the formatted product list response
type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Total      int               `json:"total"`
	Pagination PaginationInfo    `json:"pagination"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	PerPage     int `json:"per_page"`
	TotalItems  int `json:"total_items"`
}

// FormatProduct formats a product proto message into the desired response format
func FormatProduct(product *pb.Product) ProductResponse {
	// Format basic product information
	formatted := ProductResponse{
		ID:               product.Id,
		Title:            product.Title,
		Slug:             product.Slug,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		SKU:              product.Sku,
		Tags:             []string{}, // Initialize with empty array
	}

	// Format default variant ID
	if product.DefaultVariantId != nil {
		formatted.DefaultVariantID = product.DefaultVariantId.Value
	}

	// Format brand if available
	if product.Brand != nil {
		formatted.Brand = &BrandInfo{
			ID:   product.Brand.Id,
			Name: product.Brand.Name,
			Slug: product.Brand.Slug,
		}
	} else if product.BrandId != nil {
		formatted.Brand = &BrandInfo{
			ID:   product.BrandId.Value,
			Name: "Unknown", // Fallback if brand object is not populated
		}
	}

	// Format price
	currentPrices := make(map[string]float64)

	// Use the actual price from the product
	price := product.Price
	if price <= 0 && len(product.Variants) > 0 {
		// Try to get price from first variant if product price is not set
		price = product.Variants[0].Price
	}

	// Set the actual price in USD
	currentPrices["USD"] = price

	// Add EUR price as an example (in a real app, this would be converted based on exchange rates)
	currentPrices["EUR"] = price

	formatted.Price = &EnhancedPriceInfo{
		Current:  currentPrices,
		Currency: "USD", // Default currency
		Value:    price, // Add simple price value for consistency with request format
	}

	// Add discount information if available
	if product.DiscountPrice != nil && product.DiscountPrice.Value > 0 && product.Price > 0 {
		// Calculate discount percentage
		discountPercentage := ((product.Price - product.DiscountPrice.Value) / product.Price) * 100
		savingsPercentage := float64(int(discountPercentage*10)) / 10 // Round to 1 decimal place
		formatted.Price.SavingsPercentage = savingsPercentage

		// Add discount info
		formatted.Discounts = []DiscountInfo{
			{
				Type:      "percentage",
				Value:     savingsPercentage,
				ExpiresAt: formatTimestamp(nil), // Default to current time
			},
		}

		// Add the discount price to the current prices map
		currentPrices["USD_DISCOUNT"] = product.DiscountPrice.Value
		currentPrices["EUR_DISCOUNT"] = product.DiscountPrice.Value
	} else if product.Discount != nil {
		// Use discount info from product if available
		formatted.Discounts = []DiscountInfo{
			{
				Type:      product.Discount.Type,
				Value:     product.Discount.Value,
				ExpiresAt: formatTimestamp(product.Discount.ExpiresAt),
			},
		}

		// Calculate savings percentage if not already set
		if formatted.Price.SavingsPercentage == 0 && product.Discount.Type == "percentage" {
			formatted.Price.SavingsPercentage = product.Discount.Value
		}
	}

	// Format metadata
	formatted.Metadata = &MetadataInfo{
		IsPublished: product.IsPublished,
	}

	// Convert timestamps
	if product.CreatedAt != nil {
		formatted.Metadata.CreatedAt = product.CreatedAt.AsTime().Unix()
	} else {
		formatted.Metadata.CreatedAt = time.Now().Unix()
	}

	if product.UpdatedAt != nil {
		formatted.Metadata.UpdatedAt = product.UpdatedAt.AsTime().Unix()
	} else {
		formatted.Metadata.UpdatedAt = time.Now().Unix()
	}

	// Format SKU - use the provided SKU directly without modification
	formatted.SKU = product.Sku

	// Only if SKU is empty, try to get it from variants or generate one
	if formatted.SKU == "" {
		if len(product.Variants) > 0 && product.Variants[0].Sku != "" {
			// Try to use the first variant's SKU if available
			formatted.SKU = product.Variants[0].Sku
		} else {
			// Generate a SKU as last resort
			formatted.SKU = fmt.Sprintf("SKU-%s", product.Id[:8])
		}
	}

	// Format inventory
	totalQty := int(product.InventoryQty)
	status := "OUT_OF_STOCK"
	available := false

	// Use inventory status from product if available, otherwise determine from quantity
	if product.InventoryStatus != "" {
		status = product.InventoryStatus
		available = (strings.ToUpper(status) == "IN_STOCK")
	} else if totalQty > 0 {
		status = "IN_STOCK"
		available = true
	}

	// Create inventory locations from product.InventoryLocations if available
	locations := []EnhancedLocationInfo{}
	if len(product.InventoryLocations) > 0 {
		for _, loc := range product.InventoryLocations {
			locations = append(locations, EnhancedLocationInfo{
				WarehouseID: loc.WarehouseId,
				Quantity:    int(loc.AvailableQty),
			})
		}
	} else {
		// Default locations if none provided
		locations = []EnhancedLocationInfo{
			{
				WarehouseID: "A1",
				Quantity:    totalQty / 2,
			},
			{
				WarehouseID: "B2",
				Quantity:    totalQty - (totalQty / 2),
			},
		}
	}

	formatted.Inventory = &EnhancedInventoryInfo{
		Status:    status,
		Available: available,
		Quantity:  totalQty,
		Locations: locations,
	}

	// Format categories if available
	if len(product.Categories) > 0 {
		formatted.Categories = make([]CategoryInfo, len(product.Categories))
		for i, category := range product.Categories {
			formatted.Categories[i] = CategoryInfo{
				ID:   category.Id,
				Name: category.Name,
				Slug: category.Slug,
			}
		}
	}

	// Format images
	formatted.Images = make([]EnhancedImageInfo, 0, len(product.Images))
	for i, img := range product.Images {
		isThumbnail := i == 0 // First image is thumbnail by default
		viewType := "default"
		if i == 0 {
			viewType = "front"
		} else if i == 1 {
			viewType = "back"
		} else if i == 2 {
			viewType = "side"
		}

		// Ensure URL is not empty
		imgURL := img.Url
		if imgURL == "" {
			// Skip images with empty URLs
			continue
		}

		formatted.Images = append(formatted.Images, EnhancedImageInfo{
			ID:          fmt.Sprintf("img-%03d", i+1),
			URL:         imgURL,
			AltText:     img.AltText,
			Position:    int(img.Position),
			ViewType:    viewType,
			IsThumbnail: isThumbnail,
		})
	}

	// Format variants
	formatted.Variants = make([]EnhancedVariantInfo, 0, len(product.Variants))
	for _, variant := range product.Variants {
		// Skip variants with empty SKU
		if variant.Sku == "" {
			continue
		}
		formattedVariant := FormatVariant(variant)
		formatted.Variants = append(formatted.Variants, formattedVariant)
	}

	// Convert attributes if available
	if len(product.Attributes) > 0 {
		formatted.Attributes = make([]AttributeInfo, len(product.Attributes))
		for i, attr := range product.Attributes {
			formatted.Attributes[i] = AttributeInfo{
				Name:  attr.Name,
				Value: attr.Value,
			}
		}
	} else {
		formatted.Attributes = []AttributeInfo{} // Initialize with empty array
	}

	// Convert specifications if available
	if len(product.Specifications) > 0 {
		// Create array of specification objects
		formatted.Specifications = make([]SpecificationInfo, len(product.Specifications))
		for i, spec := range product.Specifications {
			formatted.Specifications[i] = SpecificationInfo{
				Name:  spec.Name,
				Value: spec.Value,
				Unit:  spec.Unit,
			}
		}
	} else {
		// Initialize with empty array
		formatted.Specifications = []SpecificationInfo{}
	}

	// Initialize reviews with enhanced structure
	formatted.Reviews = &EnhancedReviewInfo{
		Summary: EnhancedReviewSummary{
			AverageRating: 4.8, // Example rating
			TotalReviews:  127, // Example count
			RatingDistribution: map[string]int{
				"5": 98,
				"4": 20,
				"3": 5,
				"2": 2,
				"1": 2,
			},
		},
		Items: []EnhancedReviewItem{
			{
				ID:    "rev-001",
				Title: "Amazing Performance!",
				User: ReviewUser{
					ID:                "u-123",
					Name:              "John Doe",
					VerifiedPurchaser: true,
				},
				Rating:       5,
				Comment:      "Incredible performance and battery life!",
				Date:         "2024-02-20T10:00:00Z",
				HelpfulVotes: 15,
			},
		},
	}

	// Convert shipping info if available
	if product.Shipping != nil {
		formatted.Shipping = &EnhancedShippingInfo{
			FreeShipping:             product.Shipping.FreeShipping,
			EstimatedDays:            fmt.Sprintf("%d", product.Shipping.EstimatedDays),
			ExpressShippingAvailable: product.Shipping.ExpressAvailable,
		}
	} else {
		// Initialize with empty shipping info
		formatted.Shipping = &EnhancedShippingInfo{
			FreeShipping:             false,
			EstimatedDays:            "",
			ExpressShippingAvailable: false,
		}
	}

	// Convert SEO info if available
	if product.Seo != nil {
		formatted.SEO = &EnhancedSEOInfo{
			MetaTitle:       product.Seo.MetaTitle,
			MetaDescription: product.Seo.MetaDescription,
			Keywords:        product.Seo.Keywords,
			MetaTags:        product.Seo.Tags,
		}
	} else {
		// Initialize with empty SEO info
		formatted.SEO = &EnhancedSEOInfo{
			MetaTitle:       product.Title,
			MetaDescription: product.ShortDescription,
			Keywords:        []string{},
			MetaTags:        []string{},
		}
	}

	// Add tags from product tags if available
	if len(product.Tags) > 0 {
		formatted.Tags = make([]string, len(product.Tags))
		for i, tag := range product.Tags {
			formatted.Tags[i] = tag.Tag
		}
	} else {
		// Initialize with empty array
		formatted.Tags = []string{}
	}

	return formatted
}

// FormatProductList formats a list of product proto messages into the desired response format
func FormatProductList(products []*pb.Product, page, limit, total int) ProductListResponse {
	formattedProducts := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		formattedProducts = append(formattedProducts, FormatProduct(product))
	}

	totalPages := (total + limit - 1) / limit // Ceiling division

	return ProductListResponse{
		Products: formattedProducts,
		Total:    total,
		Pagination: PaginationInfo{
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     limit,
			TotalItems:  total,
		},
	}
}

// Helper function to format timestamps
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return time.Now().Format(time.RFC3339)
	}
	return ts.AsTime().Format(time.RFC3339)
}

// FormatVariant formats a variant proto message into the desired response format
func FormatVariant(variant *pb.ProductVariant) EnhancedVariantInfo {
	formattedVariant := EnhancedVariantInfo{
		ID:           variant.Id,
		ProductID:    variant.ProductId,
		SKU:          variant.Sku,
		Title:        variant.Title,
		Price:        variant.Price,
		InventoryQty: int(variant.InventoryQty),
		CreatedAt:    formatTimestamp(variant.CreatedAt),
		UpdatedAt:    formatTimestamp(variant.UpdatedAt),
	}

	// Set inherited fields
	formattedVariant.Description = variant.Description
	formattedVariant.ShortDescription = variant.ShortDescription

	// Format specifications
	if len(variant.Specifications) > 0 {
		formattedVariant.Specifications = make([]SpecificationInfo, len(variant.Specifications))
		for i, spec := range variant.Specifications {
			formattedVariant.Specifications[i] = SpecificationInfo{
				Name:  spec.Name,
				Value: spec.Value,
				Unit:  spec.Unit,
			}
		}
	}

	// Format tags
	if len(variant.Tags) > 0 {
		tags := make([]string, len(variant.Tags))
		for i, tag := range variant.Tags {
			tags[i] = tag.Tag
		}
		formattedVariant.Tags = tags
	}

	// Format categories
	if len(variant.Categories) > 0 {
		categories := make([]CategoryInfo, len(variant.Categories))
		for i, cat := range variant.Categories {
			categories[i] = CategoryInfo{
				ID:   cat.Id,
				Name: cat.Name,
				Slug: cat.Slug,
			}
		}
		formattedVariant.Categories = categories
	}

	// Format brand
	if variant.Brand != nil {
		formattedVariant.Brand = &BrandInfo{
			ID:   variant.Brand.Id,
			Name: variant.Brand.Name,
			Slug: variant.Brand.Slug,
		}
	}

	// Format SEO
	if variant.Seo != nil {
		formattedVariant.SEO = &EnhancedSEOInfo{
			MetaTitle:       variant.Seo.MetaTitle,
			MetaDescription: variant.Seo.MetaDescription,
			Keywords:        variant.Seo.Keywords,
			MetaTags:        variant.Seo.Tags,
		}
	}

	// Format shipping
	if variant.Shipping != nil {
		formattedVariant.Shipping = &EnhancedShippingInfo{
			FreeShipping:             variant.Shipping.FreeShipping,
			EstimatedDays:            fmt.Sprintf("%d days", variant.Shipping.EstimatedDays),
			ExpressShippingAvailable: variant.Shipping.ExpressAvailable,
		}
	}

	// Format discount
	if variant.Discount != nil {
		formattedVariant.Discounts = []DiscountInfo{{
			Type:      variant.Discount.Type,
			Value:     variant.Discount.Value,
			ExpiresAt: formatTimestamp(variant.Discount.ExpiresAt),
		}}
	}

	// Format attributes
	if len(variant.Attributes) > 0 {
		attributes := make([]AttributeInfo, len(variant.Attributes))
		for i, attr := range variant.Attributes {
			attributes[i] = AttributeInfo{
				Name:  attr.Name,
				Value: attr.Value,
			}
		}
		formattedVariant.Attributes = attributes
	}

	// Format images
	if len(variant.Images) > 0 {
		images := make([]EnhancedImageInfo, len(variant.Images))
		for i, img := range variant.Images {
			images[i] = EnhancedImageInfo{
				ID:       img.Id,
				URL:      img.Url,
				AltText:  img.AltText,
				Position: int(img.Position),
			}
		}
		formattedVariant.Images = images
	}

	return formattedVariant
}
