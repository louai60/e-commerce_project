# Product Service API Documentation

## Overview
The Product Service is a gRPC-based microservice that handles product management, including products, brands, categories, and images. This documentation outlines the available endpoints and their request/response structures.

## Base URL
The service is accessible via gRPC at the configured service address.

## Authentication
Authentication details should be handled at the API Gateway level.

## API Endpoints

### Products

#### 1. Create Product
- **Endpoint**: `CreateProduct`
- **Request**:
```protobuf
message CreateProductRequest {
    Product product = 1;
}
```
- **Response**: `Product` object
- **Required Fields**:
  - `title`
  - `slug`
  - `description`
  - `price`
  - `sku`
  - `inventory_qty`
  - `inventory_status`
  - `is_published`

#### 2. Get Product
- **Endpoint**: `GetProduct`
- **Request**:
```protobuf
message GetProductRequest {
    oneof identifier {
        string id = 1;
        string slug = 2;
    }
}
```
- **Response**: `Product` object
- **Notes**: Can query by either ID or slug

#### 3. List Products
- **Endpoint**: `ListProducts`
- **Request**:
```protobuf
message ListProductsRequest {
    int32 page = 1;
    int32 limit = 2;
}
```
- **Response**:
```protobuf
message ListProductsResponse {
    repeated Product products = 1;
    int32 total = 2;
}
```
- **Default Values**:
  - `page`: 1
  - `limit`: 10

#### 4. Update Product
- **Endpoint**: `UpdateProduct`
- **Request**:
```protobuf
message UpdateProductRequest {
    Product product = 1;
}
```
- **Response**: Updated `Product` object
- **Required Fields**: `id`

#### 5. Delete Product
- **Endpoint**: `DeleteProduct`
- **Request**:
```protobuf
message DeleteProductRequest {
    string id = 1;
}
```
- **Response**:
```protobuf
message DeleteProductResponse {
    bool success = 1;
}
```

### Brands

#### 1. Create Brand
- **Endpoint**: `CreateBrand`
- **Request**:
```protobuf
message CreateBrandRequest {
    Brand brand = 1;
}
```
- **Response**: `Brand` object
- **Required Fields**:
  - `name`
  - `slug`

#### 2. Get Brand
- **Endpoint**: `GetBrand`
- **Request**:
```protobuf
message GetBrandRequest {
    oneof identifier {
        string id = 1;
        string slug = 2;
    }
}
```
- **Response**: `Brand` object

#### 3. List Brands
- **Endpoint**: `ListBrands`
- **Request**:
```protobuf
message ListBrandsRequest {
    int32 page = 1;
    int32 limit = 2;
}
```
- **Response**:
```protobuf
message ListBrandsResponse {
    repeated Brand brands = 1;
    int32 total = 2;
}
```

### Categories

#### 1. Create Category
- **Endpoint**: `CreateCategory`
- **Request**:
```protobuf
message CreateCategoryRequest {
    Category category = 1;
}
```
- **Response**: `Category` object
- **Required Fields**:
  - `name`
  - `slug`

#### 2. Get Category
- **Endpoint**: `GetCategory`
- **Request**:
```protobuf
message GetCategoryRequest {
    oneof identifier {
        string id = 1;
        string slug = 2;
    }
}
```
- **Response**: `Category` object

#### 3. List Categories
- **Endpoint**: `ListCategories`
- **Request**:
```protobuf
message ListCategoriesRequest {
    int32 page = 1;
    int32 limit = 2;
}
```
- **Response**:
```protobuf
message ListCategoriesResponse {
    repeated Category categories = 1;
    int32 total = 2;
}
```

### Images

#### 1. Upload Image
- **Endpoint**: `UploadImage`
- **Request**:
```protobuf
message UploadImageRequest {
    bytes file = 1;
    string folder = 2;
    string filename = 3;
}
```
- **Response**: `UploadImageResponse` with image details

#### 2. Delete Image
- **Endpoint**: `DeleteImage`
- **Request**:
```protobuf
message DeleteImageRequest {
    string public_id = 1;
}
```
- **Response**: `DeleteImageResponse` with success status

## Data Models

### Product
```protobuf
message Product {
    string id = 1;
    string title = 2;
    string slug = 3;
    string description = 4;
    string short_description = 5;
    double price = 6;
    google.protobuf.DoubleValue discount_price = 7;
    string sku = 8;
    int32 inventory_qty = 9;
    string inventory_status = 10;
    google.protobuf.DoubleValue weight = 11;
    bool is_published = 12;
    google.protobuf.Timestamp created_at = 13;
    google.protobuf.Timestamp updated_at = 14;
    google.protobuf.StringValue brand_id = 15;
    Brand brand = 16;
    repeated ProductImage images = 17;
    repeated Category categories = 18;
    repeated ProductVariant variants = 19;
    google.protobuf.StringValue default_variant_id = 20;
    repeated ProductTag tags = 21;
    repeated ProductAttribute attributes = 22;
    repeated ProductSpecification specifications = 23;
    ProductSEO seo = 24;
    ProductShipping shipping = 25;
    ProductDiscount discount = 26;
    repeated InventoryLocation inventory_locations = 27;
}
```

### Brand
```protobuf
message Brand {
    string id = 1;
    string name = 2;
    string slug = 3;
    string description = 4;
    google.protobuf.Timestamp created_at = 5;
    google.protobuf.Timestamp updated_at = 6;
    google.protobuf.Timestamp deleted_at = 7;
}
```

### Category
```protobuf
message Category {
    string id = 1;
    string name = 2;
    string slug = 3;
    string description = 4;
    google.protobuf.StringValue parent_id = 5;
    string parent_name = 6;
    google.protobuf.Timestamp created_at = 7;
    google.protobuf.Timestamp updated_at = 8;
    google.protobuf.Timestamp deleted_at = 9;
}
```

## Error Handling
The service uses gRPC status codes for error handling:
- `INVALID_ARGUMENT`: When required fields are missing or invalid
- `NOT_FOUND`: When requested resource doesn't exist
- `INTERNAL`: For server-side errors

## Best Practices
1. Always include required fields in requests
2. Use pagination for list endpoints
3. Handle soft-deleted resources appropriately
4. Validate data before sending requests
5. Implement proper error handling for gRPC status codes

## Notes
- All timestamps are in UTC
- Soft delete is implemented for brands and categories
- Image uploads require proper file handling
- Product variants and attributes are managed through the product endpoints
- SEO, shipping, and discount information are optional fields

## Field Requirements

### Product Fields

#### Required Fields
```json
{
  "title": "Product Name",
  "slug": "product-name",
  "description": "Detailed product description",
  "price": 99.99,
  "sku": "PROD-001",
  "inventory_qty": 100,
  "inventory_status": "in_stock",
  "is_published": true
}
```

#### Optional Fields
```json
{
  "short_description": "Brief product description",
  "discount_price": 89.99,
  "weight": 1.5,
  "brand_id": "brand-uuid",
  "default_variant_id": "variant-uuid",
  "images": [
    {
      "url": "https://example.com/image.jpg",
      "alt_text": "Product image",
      "position": 1
    }
  ],
  "categories": [
    {
      "id": "category-uuid",
      "name": "Category Name",
      "slug": "category-name"
    }
  ],
  "variants": [
    {
      "sku": "VAR-001",
      "title": "Variant Title",
      "price": 99.99,
      "inventory_qty": 50,
      "attributes": [
        {
          "name": "Color",
          "value": "Red"
        }
      ]
    }
  ],
  "tags": [
    {
      "tag": "featured"
    }
  ],
  "attributes": [
    {
      "name": "Material",
      "value": "Cotton"
    }
  ],
  "specifications": [
    {
      "name": "Dimensions",
      "value": "10x20x30",
      "unit": "cm"
    }
  ],
  "seo": {
    "meta_title": "Product SEO Title",
    "meta_description": "Product SEO Description",
    "keywords": ["keyword1", "keyword2"],
    "tags": ["tag1", "tag2"]
  },
  "shipping": {
    "free_shipping": true,
    "estimated_days": 3,
    "express_available": true
  },
  "discount": {
    "type": "percentage",
    "value": 10,
    "expires_at": "2024-12-31T23:59:59Z"
  },
  "inventory_locations": [
    {
      "warehouse_id": "warehouse-uuid",
      "available_qty": 50
    }
  ]
}
```

### Brand Fields

#### Required Fields
```json
{
  "name": "Brand Name",
  "slug": "brand-name"
}
```

#### Optional Fields
```json
{
  "description": "Brand description",
  "deleted_at": null
}
```

### Category Fields

#### Required Fields
```json
{
  "name": "Category Name",
  "slug": "category-name"
}
```

#### Optional Fields
```json
{
  "description": "Category description",
  "parent_id": "parent-category-uuid",
  "parent_name": "Parent Category Name",
  "deleted_at": null
}
```

### Image Fields

#### Required Fields
```json
{
  "file": "base64_encoded_image_data",
  "folder": "products",
  "filename": "product-image.jpg"
}
```

#### Optional Fields
```json
{
  "alt_text": "Image description",
  "position": 1
}
```

## Complete JSON Examples

### Product Example
```json
{
  // Required Fields
  "title": "Premium Smartphone X",
  "slug": "premium-smartphone-x",
  "description": "Latest generation smartphone with advanced features",
  "price": 999.99,
  "sku": "PHONE-001",
  "inventory_qty": 100,
  "inventory_status": "in_stock",
  "is_published": true,

  // Optional Fields
  "short_description": "High-end smartphone with 5G capability",
  "discount_price": 899.99,
  "weight": 0.2,
  "brand_id": "brand-123e4567-e89b-12d3-a456-426614174000",
  "default_variant_id": "variant-123e4567-e89b-12d3-a456-426614174000",
  "images": [
    {
      "url": "https://example.com/images/phone-1.jpg",
      "alt_text": "Front view of Premium Smartphone X",
      "position": 1
    },
    {
      "url": "https://example.com/images/phone-2.jpg",
      "alt_text": "Back view of Premium Smartphone X",
      "position": 2
    }
  ],
  "categories": [
    {
      "id": "category-123e4567-e89b-12d3-a456-426614174000",
      "name": "Smartphones",
      "slug": "smartphones"
    }
  ],
  "variants": [
    {
      "sku": "PHONE-001-BLACK",
      "title": "Premium Smartphone X - Black",
      "price": 999.99,
      "inventory_qty": 50,
      "attributes": [
        {
          "name": "Color",
          "value": "Black"
        },
        {
          "name": "Storage",
          "value": "256GB"
        }
      ]
    }
  ],
  "tags": [
    {
      "tag": "featured"
    },
    {
      "tag": "new-arrival"
    }
  ],
  "attributes": [
    {
      "name": "Material",
      "value": "Aluminum"
    },
    {
      "name": "Screen Size",
      "value": "6.7 inches"
    }
  ],
  "specifications": [
    {
      "name": "Dimensions",
      "value": "160x75x7.5",
      "unit": "mm"
    },
    {
      "name": "Weight",
      "value": "200",
      "unit": "g"
    }
  ],
  "seo": {
    "meta_title": "Premium Smartphone X | Latest Technology",
    "meta_description": "Discover the new Premium Smartphone X with cutting-edge features",
    "keywords": ["smartphone", "5G", "premium", "latest"],
    "tags": ["mobile", "tech", "gadget"]
  },
  "shipping": {
    "free_shipping": true,
    "estimated_days": 2,
    "express_available": true
  },
  "discount": {
    "type": "percentage",
    "value": 10,
    "expires_at": "2024-12-31T23:59:59Z"
  },
  "inventory_locations": [
    {
      "warehouse_id": "warehouse-123e4567-e89b-12d3-a456-426614174000",
      "available_qty": 30
    },
    {
      "warehouse_id": "warehouse-987fcdeb-54a3-21b9-c765-987654321000",
      "available_qty": 20
    }
  ]
}
```

### Brand Example
```json
{
  // Required Fields
  "name": "TechMaster",
  "slug": "techmaster",

  // Optional Fields
  "description": "Leading technology brand since 2010",
  "deleted_at": null
}
```

### Category Example
```json
{
  // Required Fields
  "name": "Mobile Phones",
  "slug": "mobile-phones",

  // Optional Fields
  "description": "Latest mobile phones and smartphones",
  "parent_id": "category-123e4567-e89b-12d3-a456-426614174000",
  "parent_name": "Electronics",
  "deleted_at": null
}
```

### Image Upload Example
```json
{
  // Required Fields
  "file": "base64_encoded_image_data_here",
  "folder": "products/smartphones",
  "filename": "premium-phone-main.jpg",

  // Optional Fields
  "alt_text": "Main product image of Premium Smartphone X",
  "position": 1
}
```

## Notes on Field Usage
1. All IDs (product_id, brand_id, category_id, etc.) are UUID strings
2. Prices are represented as floating-point numbers
3. Timestamps should be in ISO 8601 format (e.g., "2024-01-01T00:00:00Z")
4. Boolean fields should be true/false
5. Arrays can be empty but should not be null
6. Optional fields can be omitted entirely from the request
7. Nested objects (like SEO, shipping, etc.) are completely optional 

## Variant Inheritance and Overrides

### Overview
In an eCommerce platform supporting product variants (e.g., color or size options), it's important to distinguish between data shared across all variants and data unique to each one. Proper inheritance reduces redundancy, improves data consistency, and simplifies product management.

### Current Implementation
Currently, variants have their own fields and do not automatically inherit from the parent product. Each variant must explicitly set its own values for:
- Title
- Price
- Discount Price
- Inventory Quantity
- Images
- Attributes

### Recommended Changes
The following fields should be inherited from the parent product by default, with the ability to override when necessary:

#### Inherited Fields (Default from Product)
- Description & Short Description
- Specifications
- Tags
- Categories
- Brand
- SEO Metadata
- Shipping Configuration
- Discounts: Inherit from the product unless there's a use case for per-variant discounts (e.g., specific color/size is on promotion).

#### Overrideable Fields (Set per Variant)
| Field | When to Override |
|-------|------------------|
| Price | When variant has different pricing (e.g., larger size or more storage) |
| Weight | When weight differs between variants |
| Images | For variant-specific views (e.g., different colors) |
| Title | Optional, for custom variant naming |
| Inventory Qty | Always unique per variant |

### Implementation Notes
1. **Images**: Currently, variants can have their own images. When a variant has no images, it should inherit the main product's images.
2. **Attributes**: Variants can have their own attributes to differentiate them (e.g., color, size).
3. **Inventory**: Each variant maintains its own inventory quantity.
4. **Pricing**: Variants can have different prices from the main product.

### Example Usage
```json
{
  "product": {
    "title": "Premium T-Shirt",
    "description": "High-quality cotton t-shirt",
    "price": 29.99,
    "images": [
      {
        "url": "https://example.com/tshirt-main.jpg",
        "alt_text": "Main product view"
      }
    ],
    "variants": [
      {
        "title": "Red - Large",
        "price": 29.99,
        "inventory_qty": 10,
        "attributes": [
          {
            "name": "Color",
            "value": "Red"
          },
          {
            "name": "Size",
            "value": "Large"
          }
        ],
        "images": [
          {
            "url": "https://example.com/tshirt-red.jpg",
            "alt_text": "Red variant"
          }
        ]
      },
      {
        "title": "Blue - Medium",
        "price": 29.99,
        "inventory_qty": 15,
        "attributes": [
          {
            "name": "Color",
            "value": "Blue"
          },
          {
            "name": "Size",
            "value": "Medium"
          }
        ]
        // No images specified, will inherit from product
      }
    ]
  }
}
``` 