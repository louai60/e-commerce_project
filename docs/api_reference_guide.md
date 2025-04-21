# NexCart API Reference Guide

A quick reference guide for frontend developers working with the NexCart API.

## Base URL

```
https://api.nexcart.com/api/v1
```

For local development:
```
http://localhost:8080/api/v1
```

## Authentication

Include JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

## Products API

### List Products

```
GET /products?page=1&limit=10
```

**Response:**
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
      "images": [...],
      "inventory": {
        "status": "IN_STOCK",
        "available": true,
        "quantity": 50
      }
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

```
GET /products/{id}
```
or
```
GET /products/{slug}
```

**Response:** Same as a single product in List Products

### Create Product (Admin only)

```
POST /products
```

**Request:**
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

**Response:** Created product object

### Update Product (Admin only)

```
PUT /products/{id}
```

**Request:** Same as Create Product
**Response:** Updated product object

### Delete Product (Admin only)

```
DELETE /products/{id}
```

**Response:**
```json
{
  "success": true
}
```

## User API

### Register

```
POST /users/register
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "user": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "user_type": "customer",
    "role": "user",
    "account_status": "active",
    "email_verified": false,
    "created_at": "2024-05-01T10:00:00Z",
    "updated_at": "2024-05-01T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login

```
POST /users/login
```

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
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
    "user_type": "customer",
    "role": "user",
    "account_status": "active",
    "email_verified": false,
    "last_login": "2024-05-01T15:30:00Z"
  }
}
```

### Get User Profile

```
GET /users/profile
```

**Response:**
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

```
PUT /users/profile
```

**Request:**
```json
{
  "username": "johndoe_updated",
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "1234567890"
}
```

**Response:** Updated user profile

## Brands API

### List Brands

```
GET /brands?page=1&limit=10
```

**Response:**
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

```
GET /brands/{id}
```
or
```
GET /brands/{slug}
```

**Response:**
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

### Create Brand (Admin only)

```
POST /brands
```

**Request:**
```json
{
  "brand": {
    "name": "New Brand",
    "slug": "new-brand",
    "description": "Description of the new brand"
  }
}
```

**Response:** Created brand object

## Categories API

### List Categories

```
GET /categories?page=1&limit=10
```

**Response:**
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

```
GET /categories/{id}
```
or
```
GET /categories/{slug}
```

**Response:**
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

### Create Category (Admin only)

```
POST /categories
```

**Request:**
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

**Response:** Created category object

## Error Responses

All API errors follow this format:

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
