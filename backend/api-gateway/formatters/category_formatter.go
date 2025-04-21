package formatters

import (
	pb "github.com/louai60/e-commerce_project/backend/product-service/proto"
)

// CategoryResponse represents the formatted category response
type CategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"`
	Description string  `json:"description,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	ParentName  string  `json:"parent_name,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	DeletedAt   string  `json:"deleted_at,omitempty"`
}

// CategoryListResponse represents the formatted category list response
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int                `json:"total"`
	Pagination PaginationInfo     `json:"pagination"`
}

// FormatCategory formats a category proto message into the desired response format
func FormatCategory(category *pb.Category) CategoryResponse {
	if category == nil {
		return CategoryResponse{}
	}

	formatted := CategoryResponse{
		ID:          category.Id,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ParentName:  category.ParentName,
	}

	// Format parent ID if available
	if category.ParentId != nil {
		parentID := category.ParentId.Value
		formatted.ParentID = &parentID
	}

	// Format timestamps if available
	if category.CreatedAt != nil {
		formatted.CreatedAt = formatTimestamp(category.CreatedAt)
	}

	if category.UpdatedAt != nil {
		formatted.UpdatedAt = formatTimestamp(category.UpdatedAt)
	}

	if category.DeletedAt != nil {
		formatted.DeletedAt = formatTimestamp(category.DeletedAt)
	}

	return formatted
}

// FormatCategoryList formats a list of category proto messages into the desired response format
func FormatCategoryList(categories []*pb.Category, page, limit, total int) CategoryListResponse {
	formattedCategories := make([]CategoryResponse, 0, len(categories))

	// Handle nil categories slice
	if categories != nil {
		for _, category := range categories {
			if category != nil {
				formattedCategories = append(formattedCategories, FormatCategory(category))
			}
		}
	}

	totalPages := (total + limit - 1) / limit // Ceiling division

	return CategoryListResponse{
		Categories: formattedCategories,
		Total:      total,
		Pagination: PaginationInfo{
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     limit,
			TotalItems:  total,
		},
	}
}
