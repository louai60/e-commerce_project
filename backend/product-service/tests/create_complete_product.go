package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"time"
// )

// // Request structures
// type CreateProductRequest struct {
// 	Product Product `json:"product"`
// }

// type Product struct {
// 	Title            string           `json:"title"`
// 	Slug             string           `json:"slug"`
// 	Description      string           `json:"description"`
// 	ShortDescription string           `json:"short_description"`
// 	IsPublished      bool             `json:"is_published"`
// 	InventoryStatus  string           `json:"inventory_status"`
// 	Weight           *float64         `json:"weight,omitempty"`
// 	BrandID          *StringValue     `json:"brand_id,omitempty"`
// 	Images           []ProductImage   `json:"images"`
// 	Categories       []Category       `json:"categories"`
// 	Variants         []ProductVariant `json:"variants"`
// 	Tags             []ProductTag     `json:"tags"`
// 	Attributes       []Attribute      `json:"attributes"`
// 	Specifications   []Specification  `json:"specifications"`
// 	SEO              *ProductSEO      `json:"seo,omitempty"`
// 	Shipping         *ProductShipping `json:"shipping,omitempty"`
// 	Discount         *ProductDiscount `json:"discount,omitempty"`
// }

// type StringValue struct {
// 	Value string `json:"value"`
// }

// type DoubleValue struct {
// 	Value float64 `json:"value"`
// }

// type ProductImage struct {
// 	URL      string `json:"url"`
// 	AltText  string `json:"alt_text"`
// 	Position int32  `json:"position"`
// }

// type Category struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// }

// type ProductVariant struct {
// 	Title         string             `json:"title"`
// 	SKU           string             `json:"sku"`
// 	Price         float64            `json:"price"`
// 	DiscountPrice *DoubleValue       `json:"discount_price,omitempty"`
// 	InventoryQty  int32              `json:"inventory_qty"`
// 	Attributes    []VariantAttribute `json:"attributes"`
// }

// type VariantAttribute struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// }

// type ProductTag struct {
// 	Tag string `json:"tag"`
// }

// type Attribute struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// }

// type Specification struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// 	Unit  string `json:"unit,omitempty"`
// }

// type ProductSEO struct {
// 	MetaTitle       string   `json:"meta_title"`
// 	MetaDescription string   `json:"meta_description"`
// 	Keywords        []string `json:"keywords"`
// }

// type ProductShipping struct {
// 	FreeShipping  bool    `json:"free_shipping"`
// 	Weight        float64 `json:"weight"`
// 	Dimensions    string  `json:"dimensions"`
// 	ShippingClass string  `json:"shipping_class"`
// }

// type ProductDiscount struct {
// 	Type      string  `json:"type"`
// 	Value     float64 `json:"value"`
// 	StartDate string  `json:"start_date"`
// 	EndDate   string  `json:"end_date"`
// }

// // Response structures
// type ProductResponse struct {
// 	ID               string              `json:"id"`
// 	Title            string              `json:"title"`
// 	Slug             string              `json:"slug"`
// 	ShortDescription string              `json:"short_description"`
// 	Description      string              `json:"description"`
// 	SKU              string              `json:"sku"`
// 	BrandID          string              `json:"brand_id,omitempty"`
// 	Category         *CategoryInfo       `json:"category,omitempty"`
// 	Price            *PriceInfo          `json:"price"`
// 	Inventory        *InventoryInfo      `json:"inventory"`
// 	Weight           *WeightInfo         `json:"weight,omitempty"`
// 	IsPublished      bool                `json:"is_published"`
// 	CreatedAt        string              `json:"created_at"`
// 	UpdatedAt        string              `json:"updated_at"`
// 	Images           []ImageInfo         `json:"images"`
// 	Variants         []VariantInfo       `json:"variants"`
// 	Attributes       []AttributeInfo     `json:"attributes"`
// 	Specifications   []SpecificationInfo `json:"specifications"`
// 	Tags             []string            `json:"tags"`
// }

// type CategoryInfo struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// }

// type PriceInfo struct {
// 	Amount             float64 `json:"amount"`
// 	Currency           string  `json:"currency"`
// 	OriginalAmount     float64 `json:"original_amount,omitempty"`
// 	DiscountPercentage float64 `json:"discount_percentage,omitempty"`
// }

// type InventoryInfo struct {
// 	TotalQty int    `json:"total_qty"`
// 	Status   string `json:"status"`
// }

// type WeightInfo struct {
// 	Value float64 `json:"value"`
// 	Unit  string  `json:"unit"`
// }

// type ImageInfo struct {
// 	URL string `json:"url"`
// 	Alt string `json:"alt"`
// }

// type VariantInfo struct {
// 	ID           string          `json:"id"`
// 	SKU          string          `json:"sku"`
// 	Title        string          `json:"title"`
// 	Price        float64         `json:"price"`
// 	InventoryQty int             `json:"inventory_qty"`
// 	Attributes   []AttributeInfo `json:"attributes"`
// 	CreatedAt    string          `json:"created_at"`
// 	UpdatedAt    string          `json:"updated_at"`
// }

// type AttributeInfo struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// }

// type SpecificationInfo struct {
// 	Name  string `json:"name"`
// 	Value string `json:"value"`
// 	Unit  string `json:"unit,omitempty"`
// }

// func main() {
// 	// Create a complete product
// 	productID, err := createCompleteProduct()
// 	if err != nil {
// 		log.Fatalf("Failed to create product: %v", err)
// 	}
// 	log.Printf("Successfully created product with ID: %s", productID)

// 	// Get the product to verify it was created correctly
// 	product, err := getProduct(productID)
// 	if err != nil {
// 		log.Fatalf("Failed to get product: %v", err)
// 	}

// 	// Print product details
// 	log.Printf("Retrieved product: %s", product.Title)
// 	log.Printf("Number of images: %d", len(product.Images))
// 	for i, img := range product.Images {
// 		log.Printf("Image %d: URL=%s, Alt=%s", i+1, img.URL, img.Alt)
// 	}
// 	log.Printf("Number of variants: %d", len(product.Variants))
// 	for i, variant := range product.Variants {
// 		log.Printf("Variant %d: ID=%s, SKU=%s, Title=%s", i+1, variant.ID, variant.SKU, variant.Title)
// 	}
// 	log.Printf("Number of attributes: %d", len(product.Attributes))
// 	log.Printf("Number of specifications: %d", len(product.Specifications))
// 	log.Printf("Number of tags: %d", len(product.Tags))
// }

// func createCompleteProduct() (string, error) {
// 	// Create a timestamp for unique values
// 	timestamp := time.Now().Unix()

// 	// Create the request
// 	weight := 1.5
// 	req := CreateProductRequest{
// 		Product: Product{
// 			Title:            fmt.Sprintf("Complete Test Product %d", timestamp),
// 			Slug:             fmt.Sprintf("complete-test-product-%d", timestamp),
// 			Description:      "This is a complete test product with all related data",
// 			ShortDescription: "Complete test product",
// 			IsPublished:      true,
// 			InventoryStatus:  "in_stock",
// 			Weight:           &weight,
// 			BrandID:          nil, // No brand for this test

// 			// Images
// 			Images: []ProductImage{
// 				{
// 					URL:      "https://example.com/image1.jpg",
// 					AltText:  "Test Image 1",
// 					Position: 0,
// 				},
// 				{
// 					URL:      "https://example.com/image2.jpg",
// 					AltText:  "Test Image 2",
// 					Position: 1,
// 				},
// 			},

// 			// Categories - empty for now as we don't have category IDs
// 			Categories: []Category{},

// 			// Variants
// 			Variants: []ProductVariant{
// 				{
// 					Title:        "Red Variant",
// 					SKU:          fmt.Sprintf("TEST-SKU-RED-%d", timestamp),
// 					Price:        99.99,
// 					InventoryQty: 50,
// 					Attributes: []VariantAttribute{
// 						{
// 							Name:  "Color",
// 							Value: "Red",
// 						},
// 					},
// 				},
// 				{
// 					Title:        "Blue Variant",
// 					SKU:          fmt.Sprintf("TEST-SKU-BLUE-%d", timestamp),
// 					Price:        99.99,
// 					InventoryQty: 50,
// 					Attributes: []VariantAttribute{
// 						{
// 							Name:  "Color",
// 							Value: "Blue",
// 						},
// 					},
// 				},
// 			},

// 			// Tags
// 			Tags: []ProductTag{
// 				{Tag: "test"},
// 				{Tag: "complete"},
// 				{Tag: "api"},
// 			},

// 			// Attributes
// 			Attributes: []Attribute{
// 				{
// 					Name:  "Material",
// 					Value: "Aluminum",
// 				},
// 				{
// 					Name:  "Size",
// 					Value: "Medium",
// 				},
// 			},

// 			// Specifications
// 			Specifications: []Specification{
// 				{
// 					Name:  "Dimensions",
// 					Value: "10 x 5 x 2",
// 					Unit:  "inches",
// 				},
// 				{
// 					Name:  "Weight",
// 					Value: "1.5",
// 					Unit:  "kg",
// 				},
// 			},

// 			// SEO
// 			SEO: &ProductSEO{
// 				MetaTitle:       "Complete Test Product | Your Store",
// 				MetaDescription: "This is a complete test product with all related data",
// 				Keywords:        []string{"test", "complete", "api"},
// 			},

// 			// Shipping
// 			Shipping: &ProductShipping{
// 				FreeShipping:  true,
// 				Weight:        1.5,
// 				Dimensions:    "10x5x2",
// 				ShippingClass: "standard",
// 			},

// 			// Discount
// 			Discount: &ProductDiscount{
// 				Type:      "percentage",
// 				Value:     20,
// 				StartDate: time.Now().Format(time.RFC3339),
// 				EndDate:   time.Now().AddDate(0, 1, 0).Format(time.RFC3339), // 1 month from now
// 			},
// 		},
// 	}

// 	// Convert request to JSON
// 	jsonData, err := json.MarshalIndent(req, "", "  ")
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal request: %w", err)
// 	}

// 	// Print the request for debugging
// 	log.Printf("Request JSON:\n%s", string(jsonData))

// 	// Send the request
// 	resp, err := http.Post("http://localhost:8080/api/v1/products", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to send request: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read response: %w", err)
// 	}

// 	// Check if the request was successful
// 	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	// Print the response for debugging
// 	log.Printf("Response JSON:\n%s", string(body))

// 	// Parse the response to get the product ID
// 	var response ProductResponse
// 	if err := json.Unmarshal(body, &response); err != nil {
// 		return "", fmt.Errorf("failed to unmarshal response: %w", err)
// 	}

// 	return response.ID, nil
// }

// func getProduct(productID string) (*ProductResponse, error) {
// 	// Send the request
// 	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/products/%s", productID))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to send request: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read response: %w", err)
// 	}

// 	// Check if the request was successful
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	// Parse the response
// 	var product ProductResponse
// 	if err := json.Unmarshal(body, &product); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
// 	}

// 	return &product, nil
// }
