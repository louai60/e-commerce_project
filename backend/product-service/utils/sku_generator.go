package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

// GenerateSKU creates a standardized SKU from product attributes
// Format: BRAND-CATEGORY-COLOR-SIZE-RANDOM
// Example: NIKE-SHOE-RED-42-A7X2
func GenerateSKU(brand, category, color, size string) string {
	// Sanitize and prepare each component
	brandCode := sanitizeComponent(brand, 3, 5)
	categoryCode := sanitizeComponent(category, 3, 5)
	colorCode := sanitizeComponent(color, 3, 3)
	sizeCode := sanitizeComponent(size, 1, 10) // More flexible for size

	// Generate random part (4 characters)
	randomCode := generateRandomCode(4)

	// Build the SKU with proper handling of missing components
	var skuParts []string

	// Add non-empty components
	if brandCode != "" {
		skuParts = append(skuParts, brandCode)
	}
	if categoryCode != "" {
		skuParts = append(skuParts, categoryCode)
	}
	if colorCode != "" {
		skuParts = append(skuParts, colorCode)
	}
	if sizeCode != "" {
		skuParts = append(skuParts, sizeCode)
	}

	// Always add the random part
	skuParts = append(skuParts, randomCode)

	// Join with hyphens
	return strings.Join(skuParts, "-")
}

// GenerateSKUFromProduct creates a SKU from a product struct
// It extracts brand, category, and other attributes when available
func GenerateSKUFromProduct(product *models.Product) string {
	// Extract brand name
	var brandName string
	if product.Brand != nil {
		brandName = product.Brand.Name
	}

	// Extract first category name
	var categoryName string
	if len(product.Categories) > 0 {
		categoryName = product.Categories[0].Name
	}

	// Try to find color and size from attributes
	var color, size string
	for _, attr := range product.Attributes {
		attrNameLower := strings.ToLower(attr.Name)
		if strings.Contains(attrNameLower, "color") {
			color = attr.Value
		} else if strings.Contains(attrNameLower, "size") {
			size = attr.Value
		}
	}

	// Generate the SKU
	return GenerateSKU(brandName, categoryName, color, size)
}

// SKUExistsChecker is an interface for checking if a SKU exists
type SKUExistsChecker interface {
	IsSKUExists(ctx context.Context, sku string) (bool, error)
}

// GenerateUniqueSKU creates a SKU that is guaranteed to be unique in the database
// It will try up to maxAttempts times to generate a unique SKU
func GenerateUniqueSKU(ctx context.Context, repo SKUExistsChecker, brand, category, color, size string, maxAttempts int) (string, error) {
	if maxAttempts <= 0 {
		maxAttempts = 5 // Default to 5 attempts
	}

	for i := 0; i < maxAttempts; i++ {
		// Generate a SKU
		sku := GenerateSKU(brand, category, color, size)

		// Check if it exists
		exists, err := repo.IsSKUExists(ctx, sku)
		if err != nil {
			return "", fmt.Errorf("failed to check if SKU exists: %w", err)
		}

		// If it doesn't exist, return it
		if !exists {
			return sku, nil
		}
	}

	// If we've tried maxAttempts times and still haven't found a unique SKU,
	// generate a completely random one as a last resort
	randomSKU := fmt.Sprintf("SKU-%s", generateRandomCode(8))
	return randomSKU, nil
}

// GenerateUniqueProductSKU creates a unique SKU from a product struct
func GenerateUniqueProductSKU(ctx context.Context, repo SKUExistsChecker, product *models.Product, maxAttempts int) (string, error) {
	// Extract brand name
	var brandName string
	if product.Brand != nil {
		brandName = product.Brand.Name
	}

	// Extract first category name
	var categoryName string
	if len(product.Categories) > 0 {
		categoryName = product.Categories[0].Name
	}

	// Try to find color and size from attributes
	var color, size string
	for _, attr := range product.Attributes {
		attrNameLower := strings.ToLower(attr.Name)
		if strings.Contains(attrNameLower, "color") {
			color = attr.Value
		} else if strings.Contains(attrNameLower, "size") {
			size = attr.Value
		}
	}

	// Generate a unique SKU
	return GenerateUniqueSKU(ctx, repo, brandName, categoryName, color, size, maxAttempts)
}

// sanitizeComponent prepares a component for use in a SKU
// It converts to uppercase, removes special characters, and trims to the specified length
// Returns empty string if input is empty
func sanitizeComponent(input string, minLen, maxLen int) string {
	if input == "" {
		return "" // Return empty string for missing components
	}

	// Convert to uppercase
	result := strings.ToUpper(input)

	// Remove special characters and spaces
	reg := regexp.MustCompile(`[^A-Z0-9]`)
	result = reg.ReplaceAllString(result, "")

	// Trim to length
	if len(result) < minLen {
		// Pad if too short (with X)
		for len(result) < minLen {
			result += "X"
		}
	} else if len(result) > maxLen {
		// Truncate if too long
		result = result[:maxLen]
	}

	return result
}

// generateRandomCode creates a random alphanumeric string of the specified length
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	// Generate random bytes
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to a deterministic but unique string if random generation fails
		return fmt.Sprintf("%04X", strings.ToUpper(fmt.Sprintf("%p", &length))[2:6])
	}

	// Map random bytes to charset
	for i, b := range randomBytes {
		result[i] = charset[b%byte(len(charset))]
	}

	return string(result)
}
