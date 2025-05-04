package utils

import (
	"regexp"
	"strings"
	"testing"

	"github.com/louai60/e-commerce_project/backend/product-service/models"
)

func TestGenerateSKU(t *testing.T) {
	tests := []struct {
		name      string
		brand     string
		category  string
		color     string
		size      string
		wantParts int // Number of parts expected in the SKU
	}{
		{
			name:      "Complete SKU",
			brand:     "Nike",
			category:  "Shoes",
			color:     "Red",
			size:      "42",
			wantParts: 5, // All parts present
		},
		{
			name:      "Missing Color",
			brand:     "Samsung",
			category:  "Phone",
			color:     "",
			size:      "128GB",
			wantParts: 4, // Missing color
		},
		{
			name:      "Missing Size",
			brand:     "Apple",
			category:  "Laptop",
			color:     "Silver",
			size:      "",
			wantParts: 4, // Missing size
		},
		{
			name:      "Special Characters",
			brand:     "H&M",
			category:  "T-Shirt",
			color:     "Blue/Green",
			size:      "L",
			wantParts: 5, // All parts present but sanitized
		},
		{
			name:      "Long Values",
			brand:     "VeryLongBrandName",
			category:  "ExtremelyLongCategory",
			color:     "VeryLongColorName",
			size:      "ExtraExtraLarge",
			wantParts: 5, // All parts present but truncated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSKU(tt.brand, tt.category, tt.color, tt.size)

			// Check SKU format
			parts := strings.Split(got, "-")
			if len(parts) != tt.wantParts {
				t.Errorf("GenerateSKU() returned %d parts, want %d parts. SKU: %s",
					len(parts), tt.wantParts, got)
			}

			// Check that all parts are uppercase
			for i, part := range parts {
				if part != strings.ToUpper(part) {
					t.Errorf("Part %d of SKU is not uppercase: %s", i, part)
				}
			}

			// Check that the last part is a 4-character alphanumeric code
			randomPart := parts[len(parts)-1]
			if len(randomPart) != 4 {
				t.Errorf("Random part length = %d, want 4. Part: %s", len(randomPart), randomPart)
			}

			matched, _ := regexp.MatchString("^[A-Z0-9]{4}$", randomPart)
			if !matched {
				t.Errorf("Random part is not alphanumeric: %s", randomPart)
			}
		})
	}
}

func TestGenerateSKUFromProduct(t *testing.T) {
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

	// Generate SKU from product
	sku := GenerateSKUFromProduct(product)

	// Verify SKU format
	parts := strings.Split(sku, "-")
	if len(parts) != 5 {
		t.Errorf("GenerateSKUFromProduct() returned %d parts, want 5 parts. SKU: %s",
			len(parts), sku)
	}

	// Check that brand and category are included
	if !strings.Contains(parts[0], "TEST") {
		t.Errorf("Brand part doesn't contain expected value. Got: %s", parts[0])
	}

	if !strings.Contains(parts[1], "TEST") {
		t.Errorf("Category part doesn't contain expected value. Got: %s", parts[1])
	}

	// Check color part
	if !strings.Contains(parts[2], "BLU") {
		t.Errorf("Color part doesn't contain expected value. Got: %s", parts[2])
	}

	// Check size part
	if !strings.Contains(parts[3], "MED") {
		t.Errorf("Size part doesn't contain expected value. Got: %s", parts[3])
	}
}

func TestSanitizeComponent(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		minLen int
		maxLen int
		want   string
	}{
		{
			name:   "Normal String",
			input:  "Nike",
			minLen: 3,
			maxLen: 5,
			want:   "NIKE",
		},
		{
			name:   "String with Special Characters",
			input:  "H&M Store",
			minLen: 3,
			maxLen: 5,
			want:   "HMSTO",
		},
		{
			name:   "Short String",
			input:  "HP",
			minLen: 3,
			maxLen: 5,
			want:   "HPX",
		},
		{
			name:   "Long String",
			input:  "Microsoft Corporation",
			minLen: 3,
			maxLen: 5,
			want:   "MICRO",
		},
		{
			name:   "Empty String",
			input:  "",
			minLen: 3,
			maxLen: 5,
			want:   "", // Now expecting empty string for empty input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeComponent(tt.input, tt.minLen, tt.maxLen)
			if got != tt.want {
				t.Errorf("sanitizeComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRandomCode(t *testing.T) {
	// Test multiple random codes to ensure they're different
	code1 := generateRandomCode(4)
	code2 := generateRandomCode(4)

	// Check length
	if len(code1) != 4 {
		t.Errorf("generateRandomCode() returned code of length %d, want 4", len(code1))
	}

	// Check format (alphanumeric and uppercase)
	matched, _ := regexp.MatchString("^[A-Z0-9]{4}$", code1)
	if !matched {
		t.Errorf("Random code is not alphanumeric uppercase: %s", code1)
	}

	// Check that multiple calls produce different results (this could theoretically fail, but it's very unlikely)
	if code1 == code2 {
		t.Logf("Warning: Two consecutive random codes were identical: %s. This is possible but unlikely.", code1)
	}
}
