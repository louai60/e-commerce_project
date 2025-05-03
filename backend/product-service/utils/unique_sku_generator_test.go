package utils

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

// MockSKUChecker implements the SKUExistsChecker interface for testing
type MockSKUChecker struct {
	ExistingSKUs map[string]bool
	ReturnError  bool
}

func (m *MockSKUChecker) IsSKUExists(ctx context.Context, sku string) (bool, error) {
	if m.ReturnError {
		return false, errors.New("database error")
	}
	return m.ExistingSKUs[sku], nil
}

func TestGenerateUniqueSKU(t *testing.T) {
	// Create a mock repository with some existing SKUs
	mockRepo := &MockSKUChecker{
		ExistingSKUs: map[string]bool{
			"NIKE-SHOE-RED-42-AAAA": true,
			"NIKE-SHOE-RED-42-BBBB": true,
			"NIKE-SHOE-RED-42-CCCC": true,
		},
	}

	// Test case 1: First attempt should succeed with a unique SKU
	t.Run("Unique SKU Generated", func(t *testing.T) {
		ctx := context.Background()
		sku, err := GenerateUniqueSKU(ctx, mockRepo, "Adidas", "Sneaker", "Blue", "43", 5)

		if err != nil {
			t.Errorf("GenerateUniqueSKU() error = %v, want nil", err)
			return
		}

		if sku == "" {
			t.Error("GenerateUniqueSKU() returned empty SKU")
			return
		}

		// Check SKU format
		parts := strings.Split(sku, "-")
		if len(parts) != 5 {
			t.Errorf("SKU has %d parts, want 5. SKU: %s", len(parts), sku)
		}

		// Check that brand and category are included
		if !strings.Contains(parts[0], "ADIDA") {
			t.Errorf("Brand part doesn't contain expected value. Got: %s", parts[0])
		}

		if !strings.Contains(parts[1], "SNEAK") {
			t.Errorf("Category part doesn't contain expected value. Got: %s", parts[1])
		}
	})

	// Test case 2: All attempts fail, should return a fallback SKU
	t.Run("All Attempts Fail", func(t *testing.T) {
		// Create a mock with many existing SKUs to force fallback
		alwaysExistsMock := &MockSKUChecker{
			ExistingSKUs: map[string]bool{},
		}

		// Pre-generate a bunch of SKUs to make them all "exist"
		for i := 0; i < 1000; i++ {
			alwaysExistsMock.ExistingSKUs[GenerateSKU("Nike", "Shoe", "Red", "42")] = true
		}

		ctx := context.Background()
		sku, err := GenerateUniqueSKU(ctx, alwaysExistsMock, "Nike", "Shoe", "Red", "42", 3)

		if err != nil {
			t.Errorf("GenerateUniqueSKU() error = %v, want nil", err)
			return
		}

		if sku == "" {
			t.Error("GenerateUniqueSKU() returned empty SKU")
			return
		}

		// Check that the SKU starts with "SKU-"
		if !strings.HasPrefix(sku, "SKU-") {
			t.Errorf("Fallback SKU doesn't start with 'SKU-'. Got: %s", sku)
		}

		// Check that the random part is 8 characters
		randomPart := sku[4:] // Skip "SKU-"
		if len(randomPart) != 8 {
			t.Errorf("Random part length = %d, want 8. Part: %s", len(randomPart), randomPart)
		}
	})

	// Test case 3: Repository error
	t.Run("Repository Error", func(t *testing.T) {
		// Create a mock that always returns an error
		errorMock := &MockSKUChecker{
			ReturnError: true,
		}

		ctx := context.Background()
		_, err := GenerateUniqueSKU(ctx, errorMock, "Nike", "Shoe", "Red", "42", 3)

		if err == nil {
			t.Error("GenerateUniqueSKU() error = nil, want error")
		}
	})
}

func TestGenerateUniqueProductSKU(t *testing.T) {
	// Create a mock repository
	mockRepo := &MockSKUChecker{
		ExistingSKUs: map[string]bool{
			"TESTB-TESTC-BLU-MED-AAAA": true,
		},
	}

	// Create a test product
	product := &models.Product{
		Title: "Test Product",
		Brand: &models.Brand{
			Name: "TestBrand",
		},
		Categories: []models.Category{
			{
				Name: "TestCategory",
			},
		},
		Attributes: []models.ProductAttribute{
			{
				Name:  "Color",
				Value: "Blue",
			},
			{
				Name:  "Size",
				Value: "Medium",
			},
		},
	}

	// Test case: Generate a unique SKU from a product
	t.Run("Unique Product SKU", func(t *testing.T) {
		ctx := context.Background()
		sku, err := GenerateUniqueProductSKU(ctx, mockRepo, product, 5)

		if err != nil {
			t.Errorf("GenerateUniqueProductSKU() error = %v, want nil", err)
			return
		}

		if sku == "" {
			t.Error("GenerateUniqueProductSKU() returned empty SKU")
			return
		}

		// Check that the SKU is not the one that already exists
		if sku == "TESTB-TESTC-BLU-MED-AAAA" {
			t.Errorf("Generated SKU is not unique: %s", sku)
		}

		// Check SKU format
		parts := strings.Split(sku, "-")
		if len(parts) != 5 {
			t.Errorf("SKU has %d parts, want 5. SKU: %s", len(parts), sku)
		}
	})
}
