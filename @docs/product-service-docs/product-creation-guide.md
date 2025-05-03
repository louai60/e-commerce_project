# Product Creation Guide

## Overview

This document provides guidance on how to create products in the e-commerce system. The system follows a microservices architecture with separate services for product information and inventory management.

## Product Creation JSON Structure

When creating a product, use the following JSON structure:

```json
{
  "product": {
    "title": "Product Title",
    "slug": "product-slug",
    "description": "Product description",
    "short_description": "Short description",
    "price": 99.99,
    "discount_price": 89.99,
    "sku": "PROD-001",
    "weight": 1.5,
    "is_published": true,
    "brand_id": "brand-uuid",
    "images": [
      {
        "url": "https://example.com/image.jpg",
        "alt_text": "Image description",
        "position": 1
      }
    ],
    "categories": [
      {
        "id": "category-uuid",
        "name": "Category Name"
      }
    ],
    "variants": [
      {
        "title": "Variant Title",
        "sku": "PROD-001-VAR",
        "price": 99.99,
        "discount_price": 89.99,
        "attributes": [
          {
            "name": "Color",
            "value": "Red"
          }
        ],
        "images": [
          {
            "url": "https://example.com/variant-image.jpg",
            "alt_text": "Variant image description",
            "position": 1
          }
        ]
      }
    ],
    "specifications": [
      {
        "name": "Material",
        "value": "Cotton",
        "unit": "%"
      }
    ],
    "tags": ["tag1", "tag2"],
    "seo": {
      "meta_title": "Product Meta Title",
      "meta_description": "Product meta description",
      "keywords": ["keyword1", "keyword2"],
      "meta_tags": ["tag1", "tag2"]
    },
    "shipping": {
      "free_shipping": true,
      "estimated_days": "3-5",
      "express_shipping_available": true
    },
    "discount": {
      "type": "percentage",
      "value": 10.0,
      "expires_at": "2023-12-31T23:59:59Z"
    },
    "inventory": {
      "initial_quantity": 100
    }
  }
}
```

## Important Notes

1. The `inventory` object with `initial_quantity` is used to specify the initial inventory quantity. This is the recommended approach.

2. For backward compatibility, you can also use the `inventory_qty` field, but this is deprecated and will be removed in a future version.

3. For variants, you can specify inventory quantities in the variant objects, but this is not yet fully implemented.

## Common Errors

### "column 'inventory_qty' of relation 'products' does not exist"

This error occurs when trying to create a product with the `inventory_qty` field. The field has been removed from the product service database as part of the migration to the inventory service.

**Solution**: Use the `inventory` object with `initial_quantity` instead of `inventory_qty`.

```json
{
  "product": {
    // other product fields
    "inventory": {
      "initial_quantity": 100
    }
  }
}
```

### "column 'inventory_status' of relation 'products' does not exist"

This error occurs when the system tries to access the `inventory_status` column which has been removed from the product service database.

**Solution**: This is an internal error that has been fixed. You don't need to do anything special to avoid it.

## Example: Creating a Product with Inventory

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "product": {
      "title": "Wireless Headphones",
      "slug": "wireless-headphones",
      "description": "High-quality wireless headphones",
      "short_description": "Wireless headphones with noise cancellation",
      "price": 149.99,
      "sku": "WH-001",
      "is_published": true,
      "inventory": {
        "initial_quantity": 50
      }
    }
  }'
```

## Best Practices

1. Always provide a unique SKU for each product and variant
2. Use descriptive slugs that are URL-friendly
3. Provide detailed descriptions and specifications
4. Include high-quality images with descriptive alt text
5. Specify inventory quantities accurately
6. Use SEO fields to improve search engine visibility
