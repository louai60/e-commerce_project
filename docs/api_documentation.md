# NexCart API Documentation

This document provides comprehensive documentation for the NexCart e-commerce platform API. It covers all available endpoints, request/response formats, authentication requirements, and examples for frontend developers.

## Table of Contents

1. [Base URL](#base-url)
2. [Authentication](#authentication)
3. [Error Handling](#error-handling)
4. [Products API](#products-api)
5. [User API](#user-api)
6. [Brands API](#brands-api)
7. [Categories API](#categories-api)
8. [Cart API](#cart-api)
9. [Response Formats](#response-formats)

## Base URL

All API endpoints are relative to the base URL:

```
https://api.nexcart.com/api/v1
```

For local development:

```
http://localhost:8080/api/v1
```

## Authentication

### JWT Authentication

Most endpoints require authentication using JSON Web Tokens (JWT). The token should be included in the `Authorization` header of the request:

```
Authorization: Bearer <token>
```

### Obtaining a Token

To obtain a token, use the login endpoint:

```
POST /users/login
```

The response will include:
- `token`: The JWT access token
- `refresh_token`: A refresh token for obtaining a new access token when the current one expires
- `user`: User information

### Refreshing a Token

To refresh an expired token:

```
POST /users/refresh
```

Include the refresh token in the request body:

```json
{
  "refresh_token": "<refresh_token>"
}
```

## Error Handling

All API errors follow a standard format:

```json
{
  "error": "Error message",
  "code": 400,
  "details": "Additional error details (optional)"
}
```

Common HTTP status codes:

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Products API

### List Products

Retrieves a paginated list of products.

```
GET /products
```

#### Query Parameters

| Parameter | Type   | Description                                | Default |
|-----------|--------|--------------------------------------------|---------|
| page      | int    | Page number for pagination                 | 1       |
| limit     | int    | Number of products per page                | 10      |
| category  | string | Filter by category slug                    | -       |
| price_min | float  | Minimum price filter                       | -       |
| price_max | float  | Maximum price filter                       | -       |
| sort_by   | string | Field to sort by (price, title, created_at)| -       |
| sort_order| string | Sort order (asc, desc)                     | -       |

#### Response

```json
{
  "products": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Smartphone X",
      "slug": "smartphone-x",
      "short_description": "Latest smartphone with advanced features",
      "description": "Detailed product description...",
      "sku": "PHONE-X-001",
      "price": {
        "current": {
          "USD": 999.99,
          "EUR": 849.99
        },
        "currency": "USD"
      },
      "attributes": [
        {
          "name": "Color",
          "value": "Black",
          "display_name": "Color"
        },
        {
          "name": "Storage",
          "value": "128GB",
          "display_name": "Storage Capacity"
        }
      ],
      "variants": [
        {
          "id": "variant-123",
          "sku": "PHONE-X-001-BLACK-128",
          "price": {
            "current": {
              "USD": 999.99,
              "EUR": 849.99
            },
            "currency": "USD"
          },
          "attributes": [
            {
              "name": "Color",
              "value": "Black"
            },
            {
              "name": "Storage",
              "value": "128GB"
            }
          ],
          "inventory": {
            "status": "IN_STOCK",
            "available": true,
            "quantity": 50
          }
        }
      ],
      "images": [
        {
          "id": "img-001",
          "url": "https://example.com/images/smartphone-x-1.jpg",
          "alt": "Smartphone X Front View",
          "position": 1,
          "sizes": {
            "thumbnail": "https://example.com/images/smartphone-x-1-thumb.jpg",
            "medium": "https://example.com/images/smartphone-x-1-medium.jpg",
            "large": "https://example.com/images/smartphone-x-1-large.jpg"
          }
        }
      ],
      "reviews": {
        "summary": {
          "average_rating": 4.8,
          "total_reviews": 127,
          "rating_distribution": {
            "5": 98,
            "4": 20,
            "3": 5,
            "2": 2,
            "1": 2
          }
        },
        "items": [
          {
            "id": "rev-001",
            "title": "Amazing Performance!",
            "user": {
              "id": "u-123",
              "name": "John Doe",
              "verified_purchaser": true
            },
            "rating": 5,
            "comment": "Incredible performance and battery life!",
            "date": "2024-02-20T10:00:00Z",
            "helpful_votes": 15
          }
        ]
      },
      "tags": ["smartphone", "electronics", "5G"],
      "specifications": {
        "processor": "Octa-core",
        "ram": "8GB",
        "camera": "48MP",
        "battery": "5000mAh"
      },
      "brand": {
        "id": "brand-001",
        "name": "TechBrand"
      },
      "categories": [
        {
          "id": "cat-001",
          "name": "Smartphones",
          "slug": "smartphones"
        }
      ],
      "inventory": {
        "status": "IN_STOCK",
        "available": true,
        "quantity": 50,
        "locations": [
          {
            "warehouse_id": "A1",
            "quantity": 25
          },
          {
            "warehouse_id": "B2",
            "quantity": 25
          }
        ]
      },
      "metadata": {
        "created_at": 1645347600,
        "updated_at": 1645347600
      },
      "seo": {
        "meta_title": "Smartphone X - Latest Technology",
        "meta_description": "Discover the amazing Smartphone X with advanced features and technology.",
        "keywords": ["smartphone", "mobile phone", "5G"]
      },
      "shipping": {
        "free_shipping": true,
        "estimated_days": "2-3",
        "express_shipping_available": true,
        "express_shipping_days": "1"
      },
      "discounts": [
        {
          "type": "PERCENTAGE",
          "value": 10,
          "expires_at": "2024-12-31T23:59:59Z"
        }
      ]
    }
  ],
  "total": 100,
  "pagination": {
    "current_page": 1,
    "total_pages": 10,
    "per_page": 10,
    "total_items": 100
  }
}
```

### Get Product

Retrieves detailed information about a specific product.

```
GET /products/{id}
```

or

```
GET /products/{slug}
```

#### Response

The response format is the same as a single product in the List Products endpoint.

### Create Product

Creates a new product. Requires admin authentication.

```
POST /products
```

#### Request Body

```json
{
  "product": {
    "title": "New Smartphone",
    "slug": "new-smartphone",
    "description": "Detailed product description...",
    "short_description": "Brief product description",
    "price": 899.99,
    "sku": "PHONE-NEW-001",
    "inventory_qty": 100,
    "inventory_status": "in_stock",
    "is_published": true,
    "brand_id": "brand-001",
    "images": [
      {
        "url": "https://example.com/images/new-smartphone-1.jpg",
        "alt": "New Smartphone Front View",
        "position": 1
      }
    ]
  }
}
```

#### Response

Returns the created product in the same format as Get Product.

### Update Product

Updates an existing product. Requires admin authentication.

```
PUT /products/{id}
```

#### Request Body

Same format as Create Product.

#### Response

Returns the updated product in the same format as Get Product.

### Delete Product

Deletes a product. Requires admin authentication.

```
DELETE /products/{id}
```

#### Response

```json
{
  "success": true
}
```

## User API

### Register

Creates a new user account.

```
POST /users/register
```

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Response

```json
{
  "user": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "",
    "user_type": "customer",
    "role": "user",
    "account_status": "active",
    "email_verified": false,
    "phone_verified": false,
    "created_at": "2024-05-01T10:00:00Z",
    "updated_at": "2024-05-01T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login

Authenticates a user and returns tokens.

```
POST /users/login
```

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

#### Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "",
    "user_type": "customer",
    "role": "user",
    "account_status": "active",
    "email_verified": false,
    "phone_verified": false,
    "created_at": "2024-05-01T10:00:00Z",
    "updated_at": "2024-05-01T10:00:00Z",
    "last_login": "2024-05-01T15:30:00Z"
  }
}
```

### Get User Profile

Retrieves the authenticated user's profile.

```
GET /users/profile
```

#### Response

```json
{
  "user": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "phone_number": "",
    "user_type": "customer",
    "role": "user",
    "account_status": "active",
    "email_verified": false,
    "phone_verified": false,
    "created_at": "2024-05-01T10:00:00Z",
    "updated_at": "2024-05-01T10:00:00Z",
    "last_login": "2024-05-01T15:30:00Z"
  }
}
```

### Update User Profile

Updates the authenticated user's profile.

```
PUT /users/profile
```

#### Request Body

```json
{
  "username": "johndoe_updated",
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "1234567890"
}
```

#### Response

Returns the updated user profile in the same format as Get User Profile.

### Add Address

Adds a new address to the authenticated user's profile.

```
POST /users/addresses
```

#### Request Body

```json
{
  "address_line1": "123 Main St",
  "address_line2": "Apt 4B",
  "city": "New York",
  "state": "NY",
  "postal_code": "10001",
  "country": "USA",
  "is_default": true,
  "address_type": "shipping"
}
```

#### Response

```json
{
  "address_id": "addr-001",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "address_line1": "123 Main St",
  "address_line2": "Apt 4B",
  "city": "New York",
  "state": "NY",
  "postal_code": "10001",
  "country": "USA",
  "is_default": true,
  "address_type": "shipping",
  "created_at": "2024-05-01T10:00:00Z",
  "updated_at": "2024-05-01T10:00:00Z"
}
```

### Add Payment Method

Adds a new payment method to the authenticated user's profile.

```
POST /users/payment-methods
```

#### Request Body

```json
{
  "card_number": "4111111111111111",
  "card_holder_name": "John Doe",
  "expiry_month": 12,
  "expiry_year": 2025,
  "card_type": "visa",
  "is_default": true
}
```

#### Response

```json
{
  "payment_method_id": "pm-001",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "card_last_four": "1111",
  "card_holder_name": "John Doe",
  "expiry_month": 12,
  "expiry_year": 2025,
  "card_type": "visa",
  "is_default": true,
  "created_at": "2024-05-01T10:00:00Z",
  "updated_at": "2024-05-01T10:00:00Z"
}
```

## Brands API

### List Brands

Retrieves a paginated list of brands.

```
GET /brands
```

#### Query Parameters

| Parameter | Type | Description                | Default |
|-----------|------|----------------------------|---------|
| page      | int  | Page number for pagination | 1       |
| limit     | int  | Number of brands per page  | 10      |

#### Response

```json
{
  "brands": [
    {
      "id": "brand-001",
      "name": "TechBrand",
      "slug": "tech-brand",
      "description": "Leading technology brand"
    }
  ],
  "total": 50,
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "per_page": 10,
    "total_items": 50
  }
}
```

### Get Brand

Retrieves detailed information about a specific brand.

```
GET /brands/{id}
```

or

```
GET /brands/{slug}
```

#### Response

```json
{
  "id": "brand-001",
  "name": "TechBrand",
  "slug": "tech-brand",
  "description": "Leading technology brand",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Create Brand

Creates a new brand. Requires admin authentication.

```
POST /brands
```

#### Request Body

```json
{
  "brand": {
    "name": "New Brand",
    "slug": "new-brand",
    "description": "Description of the new brand"
  }
}
```

#### Response

Returns the created brand in the same format as Get Brand.

## Categories API

### List Categories

Retrieves a paginated list of categories.

```
GET /categories
```

#### Query Parameters

| Parameter | Type | Description                   | Default |
|-----------|------|-------------------------------|---------|
| page      | int  | Page number for pagination    | 1       |
| limit     | int  | Number of categories per page | 10      |

#### Response

```json
{
  "categories": [
    {
      "id": "cat-001",
      "name": "Electronics",
      "slug": "electronics",
      "description": "Electronic devices and gadgets",
      "parent_id": null,
      "parent_name": null
    },
    {
      "id": "cat-002",
      "name": "Smartphones",
      "slug": "smartphones",
      "description": "Mobile phones and accessories",
      "parent_id": "cat-001",
      "parent_name": "Electronics"
    }
  ],
  "total": 30,
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "per_page": 10,
    "total_items": 30
  }
}
```

### Get Category

Retrieves detailed information about a specific category.

```
GET /categories/{id}
```

or

```
GET /categories/{slug}
```

#### Response

```json
{
  "id": "cat-001",
  "name": "Electronics",
  "slug": "electronics",
  "description": "Electronic devices and gadgets",
  "parent_id": null,
  "parent_name": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Create Category

Creates a new category. Requires admin authentication.

```
POST /categories
```

#### Request Body

```json
{
  "category": {
    "name": "New Category",
    "slug": "new-category",
    "description": "Description of the new category",
    "parent_id": "cat-001"
  }
}
```

#### Response

Returns the created category in the same format as Get Category.

## Cart API

The cart is managed client-side using Redux. The cart data structure is as follows:

```typescript
type CartItem = {
  id: number;
  title: string;
  price: number;
  discountedPrice: number;
  quantity: number;
  imgs?: {
    thumbnails: string[];
    previews: string[];
  };
};
```

### Cart Operations

The following Redux actions are available for cart management:

- `addItemToCart`: Adds an item to the cart
- `removeItemFromCart`: Removes an item from the cart
- `updateCartItemQuantity`: Updates the quantity of an item in the cart
- `removeAllItemsFromCart`: Clears the cart

## Response Formats

### Product Response Format

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Smartphone X",
  "slug": "smartphone-x",
  "short_description": "Latest smartphone with advanced features",
  "description": "Detailed product description...",
  "sku": "PHONE-X-001",
  "price": {
    "current": {
      "USD": 999.99,
      "EUR": 849.99
    },
    "currency": "USD"
  },
  "attributes": [
    {
      "name": "Color",
      "value": "Black",
      "display_name": "Color"
    },
    {
      "name": "Storage",
      "value": "128GB",
      "display_name": "Storage Capacity"
    }
  ],
  "variants": [
    {
      "id": "variant-123",
      "sku": "PHONE-X-001-BLACK-128",
      "price": {
        "current": {
          "USD": 999.99,
          "EUR": 849.99
        },
        "currency": "USD"
      },
      "attributes": [
        {
          "name": "Color",
          "value": "Black"
        },
        {
          "name": "Storage",
          "value": "128GB"
        }
      ],
      "inventory": {
        "status": "IN_STOCK",
        "available": true,
        "quantity": 50
      }
    }
  ],
  "images": [
    {
      "id": "img-001",
      "url": "https://example.com/images/smartphone-x-1.jpg",
      "alt": "Smartphone X Front View",
      "position": 1,
      "sizes": {
        "thumbnail": "https://example.com/images/smartphone-x-1-thumb.jpg",
        "medium": "https://example.com/images/smartphone-x-1-medium.jpg",
        "large": "https://example.com/images/smartphone-x-1-large.jpg"
      }
    }
  ],
  "reviews": {
    "summary": {
      "average_rating": 4.8,
      "total_reviews": 127,
      "rating_distribution": {
        "5": 98,
        "4": 20,
        "3": 5,
        "2": 2,
        "1": 2
      }
    },
    "items": [
      {
        "id": "rev-001",
        "title": "Amazing Performance!",
        "user": {
          "id": "u-123",
          "name": "John Doe",
          "verified_purchaser": true
        },
        "rating": 5,
        "comment": "Incredible performance and battery life!",
        "date": "2024-02-20T10:00:00Z",
        "helpful_votes": 15
      }
    ]
  },
  "tags": ["smartphone", "electronics", "5G"],
  "specifications": {
    "processor": "Octa-core",
    "ram": "8GB",
    "camera": "48MP",
    "battery": "5000mAh"
  },
  "brand": {
    "id": "brand-001",
    "name": "TechBrand"
  },
  "categories": [
    {
      "id": "cat-001",
      "name": "Smartphones",
      "slug": "smartphones"
    }
  ],
  "inventory": {
    "status": "IN_STOCK",
    "available": true,
    "quantity": 50,
    "locations": [
      {
        "warehouse_id": "A1",
        "quantity": 25
      },
      {
        "warehouse_id": "B2",
        "quantity": 25
      }
    ]
  },
  "metadata": {
    "created_at": 1645347600,
    "updated_at": 1645347600
  },
  "seo": {
    "meta_title": "Smartphone X - Latest Technology",
    "meta_description": "Discover the amazing Smartphone X with advanced features and technology.",
    "keywords": ["smartphone", "mobile phone", "5G"]
  },
  "shipping": {
    "free_shipping": true,
    "estimated_days": "2-3",
    "express_shipping_available": true,
    "express_shipping_days": "1"
  },
  "discounts": [
    {
      "type": "PERCENTAGE",
      "value": 10,
      "expires_at": "2024-12-31T23:59:59Z"
    }
  ]
}
```

### User Response Format

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "username": "johndoe",
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "",
  "user_type": "customer",
  "role": "user",
  "account_status": "active",
  "email_verified": false,
  "phone_verified": false,
  "created_at": "2024-05-01T10:00:00Z",
  "updated_at": "2024-05-01T10:00:00Z",
  "last_login": "2024-05-01T15:30:00Z"
}
```

### Pagination Format

```json
{
  "current_page": 1,
  "total_pages": 10,
  "per_page": 10,
  "total_items": 100
}
```
