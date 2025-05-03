# Inventory Service Migration

This document outlines the migration of inventory-related functionality from the product service to the dedicated inventory service.

## Overview

As part of our microservices architecture, we've moved inventory management from the product service to a dedicated inventory service. This separation of concerns allows for better scalability, maintainability, and domain isolation.

## Changes Made

### 1. Database Schema Changes

We've created a migration (`000013_remove_inventory_fields.up.sql`) that removes inventory-related fields from the product service database:

- Removed from `products` table:
  - `inventory_qty` - Now managed by inventory service
  - `inventory_status` - Now derived from inventory service data

- Removed from `product_variants` table:
  - `inventory_qty` - Now managed by inventory service

- Dropped the `product_inventory_locations` table - Now managed by inventory service

### 2. Model Changes

We've updated the product models to reflect these changes:

- Removed `InventoryQty` and `InventoryStatus` fields from the `Product` struct
- Removed `InventoryQty` field from the `ProductVariant` struct
- Removed `InventoryLocations` field from the `Product` struct
- Kept `SKU` fields as reference keys to the inventory service

### 3. Integration with Inventory Service

We've implemented integration between the product service and inventory service:

- Added an inventory client in the product service
- Updated the product creation flow to create inventory items in the inventory service
- Added inventory availability checking during product operations

## Data Flow

### Product Creation

1. Product service creates the product record in its database
2. Product service calls the inventory service to create an inventory item
3. Inventory service stores all inventory-related data

### Inventory Retrieval

1. When product data is requested, the product service retrieves core product information
2. If inventory information is needed, the product service calls the inventory service
3. The inventory service returns the current inventory status and quantity

## Benefits

This migration provides several benefits:

1. **Clear Domain Separation**: Each service is responsible for its own domain
2. **Improved Scalability**: Services can be scaled independently based on load
3. **Better Maintainability**: Changes to inventory logic don't affect product logic
4. **Enhanced Data Consistency**: Inventory data is managed in one place

## Future Considerations

1. **Caching**: Implement caching of frequently accessed inventory data
2. **Event-Based Communication**: Consider using events for asynchronous updates
3. **Fallback Mechanisms**: Implement fallbacks for when the inventory service is unavailable
