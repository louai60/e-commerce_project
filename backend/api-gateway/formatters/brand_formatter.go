package formatters

import (
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

// BrandResponse represents the formatted brand response
type BrandResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	DeletedAt   string `json:"deleted_at,omitempty"`
}

// BrandListResponse represents the formatted brand list response
type BrandListResponse struct {
	Brands     []BrandResponse `json:"brands"`
	Total      int             `json:"total"`
	Pagination PaginationInfo  `json:"pagination"`
}

// FormatBrand formats a brand proto message into the desired response format
func FormatBrand(brand *pb.Brand) BrandResponse {
	if brand == nil {
		return BrandResponse{}
	}

	formatted := BrandResponse{
		ID:          brand.Id,
		Name:        brand.Name,
		Slug:        brand.Slug,
		Description: brand.Description,
	}

	// Format timestamps if available
	if brand.CreatedAt != nil {
		formatted.CreatedAt = formatTimestamp(brand.CreatedAt)
	}

	if brand.UpdatedAt != nil {
		formatted.UpdatedAt = formatTimestamp(brand.UpdatedAt)
	}

	if brand.DeletedAt != nil {
		formatted.DeletedAt = formatTimestamp(brand.DeletedAt)
	}

	return formatted
}

// FormatBrandList formats a list of brand proto messages into the desired response format
func FormatBrandList(brands []*pb.Brand, page, limit, total int) BrandListResponse {
	formattedBrands := make([]BrandResponse, 0, len(brands))

	// Handle nil brands slice
	if brands != nil {
		for _, brand := range brands {
			if brand != nil {
				formattedBrands = append(formattedBrands, FormatBrand(brand))
			}
		}
	}

	totalPages := (total + limit - 1) / limit // Ceiling division

	return BrandListResponse{
		Brands: formattedBrands,
		Total:  total,
		Pagination: PaginationInfo{
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     limit,
			TotalItems:  total,
		},
	}
}
