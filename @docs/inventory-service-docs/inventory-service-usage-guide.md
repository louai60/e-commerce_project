# Inventory Service Usage Guide

## Overview

This document provides guidance on how to use the inventory service in conjunction with the product service. The inventory service is responsible for managing all inventory-related operations, while the product service focuses on product information.

## Architecture

The system follows a microservices architecture with clear domain separation:

1. **Product Service**: Manages product information (title, description, price, etc.)
2. **Inventory Service**: Manages inventory information (quantities, locations, etc.)

## Creating Products with Inventory

When creating a product, you need to:

1. Create the product in the product service
2. Create the inventory item in the inventory service

### JSON Structure for Product Creation

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
    "inventory": {
      "initial_quantity": 100
    }
  }
}
```

### Important Notes

1. The `inventory_qty` field has been removed from the product service database. Instead, use the `inventory` object with `initial_quantity` to specify the initial inventory quantity.

2. The API gateway will automatically create an inventory item in the inventory service when a product is created with the `inventory` object.

3. For variants, inventory is managed separately in the inventory service. Each variant can have its own inventory record.

## Retrieving Inventory Information

To get inventory information for a product:

1. Retrieve the product from the product service
2. Use the product ID to retrieve inventory information from the inventory service

## Updating Inventory

Inventory updates should be performed through the inventory service API. The product service should not be used to update inventory quantities.

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

## Best Practices

1. Always use the inventory service for inventory-related operations
2. Keep product and inventory concerns separate
3. Use the `inventory` object when creating products with initial inventory
4. For bulk operations, consider using batch APIs provided by the inventory service
