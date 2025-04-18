# Product Service Implementation Plan

This document outlines the step-by-step plan to fully implement the product service's related data functionality. Each section represents a related data type that needs to be implemented, with specific tasks for each.

## Table of Contents
- [Database Tables](#database-tables)
- [Implementation Steps](#implementation-steps)
  - [1. Product Images](#1-product-images)
  - [2. Product Categories](#2-product-categories)
  - [3. Product Variants](#3-product-variants)
  - [4. Product Tags](#4-product-tags)
  - [5. Product Attributes](#5-product-attributes)
  - [6. Product Specifications](#6-product-specifications)
  - [7. Product SEO](#7-product-seo)
  - [8. Product Shipping](#8-product-shipping)
  - [9. Product Discounts](#9-product-discounts)
  - [10. Inventory Locations](#10-inventory-locations)
- [Testing Plan](#testing-plan)
- [Deployment Checklist](#deployment-checklist)

## Database Tables

The following tables need to be created or verified in the database:

- [x] `product_images`
- [ ] `product_categories` (junction table)
- [ ] `categories` (if not exists)
- [ ] `product_variants`
- [ ] `variant_attributes`
- [ ] `product_tags`
- [ ] `product_attributes`
- [ ] `product_specifications`
- [ ] `product_seo`
- [ ] `product_shipping`
- [ ] `product_discounts`
- [ ] `inventory_locations`
- [ ] `warehouses` (if not exists)

## Implementation Steps

### 1. Product Images

- [x] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_images (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      url TEXT NOT NULL,
      alt_text TEXT,
      position INTEGER DEFAULT 0,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
  );
  ```
  ✅ Table already exists in the database. We added the missing `created_at` and `updated_at` columns.

- [x] **Update Repository**
  - Add methods to create, read, update, and delete product images
  - Implement batch operations for multiple images
  ✅ Repository methods already implemented in `postgres_repository.go` and `postgres/product_repository.go`

- [x] **Update Service Layer**
  - Modify the product service to handle image operations
  - Ensure images are properly processed during product creation and updates
  ✅ Service layer already handles images in `product_service.go`

- [x] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool
  ✅ Functionality verified through manual testing and a custom test script

### 2. Product Categories

- [ ] **Create/Verify Database Tables**
  ```sql
  -- If not exists
  CREATE TABLE IF NOT EXISTS categories (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name TEXT NOT NULL,
      slug TEXT NOT NULL UNIQUE,
      description TEXT,
      parent_id UUID REFERENCES categories(id),
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      deleted_at TIMESTAMP WITH TIME ZONE
  );

  -- Junction table
  CREATE TABLE IF NOT EXISTS product_categories (
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      PRIMARY KEY (product_id, category_id)
  );
  ```

- [ ] **Update Repository**
  - Add methods to associate products with categories
  - Implement category retrieval for products

- [ ] **Update Service Layer**
  - Modify the product service to handle category associations
  - Ensure categories are properly processed during product creation and updates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 3. Product Variants

- [ ] **Create/Verify Database Tables**
  ```sql
  CREATE TABLE IF NOT EXISTS product_variants (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      sku TEXT,
      price DECIMAL(10, 2) NOT NULL,
      discount_price DECIMAL(10, 2),
      inventory_qty INTEGER NOT NULL DEFAULT 0,
      inventory_status TEXT,
      weight DECIMAL(10, 2),
      is_default BOOLEAN DEFAULT FALSE,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
  );

  CREATE TABLE IF NOT EXISTS variant_attributes (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
      name TEXT NOT NULL,
      value TEXT NOT NULL,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product variants
  - Implement variant attribute operations
  - Add methods to set default variant

- [ ] **Update Service Layer**
  - Modify the product service to handle variant operations
  - Ensure variants are properly processed during product creation and updates
  - Implement default variant logic

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 4. Product Tags

- [ ] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_tags (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      tag TEXT NOT NULL,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id, tag)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product tags
  - Implement batch operations for multiple tags

- [ ] **Update Service Layer**
  - Modify the product service to handle tag operations
  - Ensure tags are properly processed during product creation and updates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 5. Product Attributes

- [ ] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_attributes (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      name TEXT NOT NULL,
      value TEXT NOT NULL,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id, name)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product attributes
  - Implement batch operations for multiple attributes

- [ ] **Update Service Layer**
  - Modify the product service to handle attribute operations
  - Ensure attributes are properly processed during product creation and updates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 6. Product Specifications

- [ ] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_specifications (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      name TEXT NOT NULL,
      value TEXT NOT NULL,
      unit TEXT,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id, name)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product specifications
  - Implement batch operations for multiple specifications

- [ ] **Update Service Layer**
  - Modify the product service to handle specification operations
  - Ensure specifications are properly processed during product creation and updates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 7. Product SEO

- [ ] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_seo (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      meta_title TEXT,
      meta_description TEXT,
      keywords TEXT[],
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product SEO information

- [ ] **Update Service Layer**
  - Modify the product service to handle SEO operations
  - Ensure SEO data is properly processed during product creation and updates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 8. Product Shipping

- [ ] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_shipping (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      free_shipping BOOLEAN DEFAULT FALSE,
      weight DECIMAL(10, 2),
      dimensions TEXT,
      shipping_class TEXT,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product shipping information

- [ ] **Update Service Layer**
  - Modify the product service to handle shipping operations
  - Ensure shipping data is properly processed during product creation and updates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 9. Product Discounts

- [ ] **Create/Verify Database Table**
  ```sql
  CREATE TABLE IF NOT EXISTS product_discounts (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      type TEXT NOT NULL, -- 'percentage', 'fixed', etc.
      value DECIMAL(10, 2) NOT NULL,
      start_date TIMESTAMP WITH TIME ZONE,
      end_date TIMESTAMP WITH TIME ZONE,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete product discount information

- [ ] **Update Service Layer**
  - Modify the product service to handle discount operations
  - Ensure discount data is properly processed during product creation and updates
  - Implement logic to check if a discount is active based on dates

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

### 10. Inventory Locations

- [ ] **Create/Verify Database Tables**
  ```sql
  -- If not exists
  CREATE TABLE IF NOT EXISTS warehouses (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name TEXT NOT NULL,
      address TEXT,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
  );

  CREATE TABLE IF NOT EXISTS product_inventory_locations (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
      available_qty INTEGER NOT NULL DEFAULT 0,
      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
      UNIQUE (product_id, warehouse_id)
  );
  ```

- [ ] **Update Repository**
  - Add methods to create, read, update, and delete inventory location information
  - Implement batch operations for multiple locations

- [ ] **Update Service Layer**
  - Modify the product service to handle inventory location operations
  - Ensure inventory location data is properly processed during product creation and updates
  - Implement logic to calculate total inventory across locations

- [ ] **Test Implementation**
  - Write unit tests for repository methods
  - Write integration tests for the service layer
  - Test API endpoints with Postman or similar tool

## Testing Plan

- [ ] **Unit Tests**
  - Repository methods for each related data type
  - Service layer methods for each related data type

- [ ] **Integration Tests**
  - End-to-end tests for product creation with all related data
  - End-to-end tests for product updates with all related data
  - End-to-end tests for product retrieval with all related data

- [ ] **API Tests**
  - Create Postman collection for testing all endpoints
  - Test all CRUD operations for products and related data

## Deployment Checklist

- [ ] **Database Migrations**
  - Create migration scripts for all new tables
  - Test migrations in development environment
  - Plan for production deployment

- [ ] **Service Updates**
  - Update service configuration if needed
  - Update API documentation
  - Update client libraries if needed

- [ ] **Monitoring**
  - Add logging for new operations
  - Set up alerts for potential issues
  - Update dashboards to include new metrics

- [ ] **Documentation**
  - Update API documentation
  - Update internal documentation
  - Create examples for common operations
