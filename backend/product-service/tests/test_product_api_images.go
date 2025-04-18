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

// // Define structs to match the API request/response format
// type ProductImage struct {
// 	URL      string `json:"url"`
// 	AltText  string `json:"alt_text"`
// 	Position int32  `json:"position"`
// }

// type ProductVariant struct {
// 	Title        string  `json:"title"`
// 	SKU          string  `json:"sku"`
// 	Price        float64 `json:"price"`
// 	InventoryQty int32   `json:"inventory_qty"`
// }

// type ProductRequest struct {
// 	Product struct {
// 		Title            string           `json:"title"`
// 		Slug             string           `json:"slug"`
// 		Description      string           `json:"description"`
// 		ShortDescription string           `json:"short_description"`
// 		IsPublished      bool             `json:"is_published"`
// 		InventoryStatus  string           `json:"inventory_status"`
// 		Images           []ProductImage   `json:"images"`
// 		Variants         []ProductVariant `json:"variants"`
// 	} `json:"product"`
// }

// type ProductResponse struct {
// 	ID               string        `json:"id"`
// 	Title            string        `json:"title"`
// 	Slug             string        `json:"slug"`
// 	ShortDescription string        `json:"short_description"`
// 	Description      string        `json:"description"`
// 	IsPublished      bool          `json:"is_published"`
// 	CreatedAt        string        `json:"created_at"`
// 	UpdatedAt        string        `json:"updated_at"`
// 	Images           []ImageInfo   `json:"images"`
// 	Variants         []VariantInfo `json:"variants"`
// }

// type ImageInfo struct {
// 	URL string `json:"url"`
// 	Alt string `json:"alt"`
// }

// type VariantInfo struct {
// 	ID           string  `json:"id"`
// 	SKU          string  `json:"sku"`
// 	Title        string  `json:"title"`
// 	Price        float64 `json:"price"`
// 	InventoryQty int     `json:"inventory_qty"`
// }

// func main() {
// 	// 1. Create a product with images
// 	productID, err := createProductWithImages()
// 	if err != nil {
// 		log.Fatalf("Failed to create product: %v", err)
// 	}
// 	log.Printf("Created product with ID: %s", productID)

// 	// 2. Retrieve the product to check if images are included
// 	product, err := getProduct(productID)
// 	if err != nil {
// 		log.Fatalf("Failed to get product: %v", err)
// 	}

// 	// 3. Print the product details
// 	log.Printf("Retrieved product: %s", product.Product.Title)
// 	log.Printf("Number of images: %d", len(product.Product.Images))
// 	for i, img := range product.Product.Images {
// 		log.Printf("Image %d: URL=%s, AltText=%s", i+1, img.URL, img.AltText)
// 	}
// }

// func createProductWithImages() (string, error) {
// 	// Create the request body
// 	var req ProductRequest
// 	req.Product.Title = "Test Product with Images API"
// 	req.Product.Slug = fmt.Sprintf("test-product-images-api-%d", time.Now().Unix())
// 	req.Product.Description = "This is a test product with images via API"
// 	req.Product.ShortDescription = "Test product with images via API"
// 	req.Product.IsPublished = true
// 	req.Product.InventoryStatus = "in_stock"

// 	// Add images
// 	req.Product.Images = []ProductImage{
// 		{
// 			URL:      "https://example.com/image1.jpg",
// 			AltText:  "Test Image 1",
// 			Position: 0,
// 		},
// 		{
// 			URL:      "https://example.com/image2.jpg",
// 			AltText:  "Test Image 2",
// 			Position: 1,
// 		},
// 	}

// 	// Add a variant
// 	req.Product.Variants = []ProductVariant{
// 		{
// 			Title:        "Test Variant",
// 			SKU:          fmt.Sprintf("TEST-SKU-%d", time.Now().Unix()),
// 			Price:        99.99,
// 			InventoryQty: 100,
// 		},
// 	}

// 	// Convert request to JSON
// 	jsonData, err := json.Marshal(req)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal request: %w", err)
// 	}

// 	// Send the request
// 	resp, err := http.Post("http://localhost:8080/api/products", "application/json", bytes.NewBuffer(jsonData))
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
// 	if resp.StatusCode != http.StatusCreated {
// 		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
// 	}

// 	// Parse the response to get the product ID
// 	var response map[string]interface{}
// 	if err := json.Unmarshal(body, &response); err != nil {
// 		return "", fmt.Errorf("failed to unmarshal response: %w", err)
// 	}

// 	// Extract the product ID
// 	productID, ok := response["id"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("failed to extract product ID from response")
// 	}

// 	return productID, nil
// }

// func getProduct(productID string) (*ProductRequest, error) {
// 	// Send the request
// 	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/products/%s", productID))
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
// 	var product ProductRequest
// 	if err := json.Unmarshal(body, &product); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
// 	}

// 	return &product, nil
// }
