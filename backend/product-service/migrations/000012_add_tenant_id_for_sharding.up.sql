-- Migration: 000012_add_tenant_id_for_sharding (Up)

-- Step 1: Add tenant_id column to products table
ALTER TABLE products
ADD COLUMN tenant_id VARCHAR(50) DEFAULT NULL;

-- Step 2: Add tenant_id column to brands table
ALTER TABLE brands
ADD COLUMN tenant_id VARCHAR(50) DEFAULT NULL;

-- Step 3: Add tenant_id column to categories table
ALTER TABLE categories
ADD COLUMN tenant_id VARCHAR(50) DEFAULT NULL;

-- Step 4: Create indexes on tenant_id columns for faster lookups
CREATE INDEX idx_products_tenant_id ON products(tenant_id);
CREATE INDEX idx_brands_tenant_id ON brands(tenant_id);
CREATE INDEX idx_categories_tenant_id ON categories(tenant_id);

-- Step 5: Create composite indexes for tenant_id and other common query fields
CREATE INDEX idx_products_tenant_id_created_at ON products(tenant_id, created_at);
CREATE INDEX idx_products_tenant_id_is_published ON products(tenant_id, is_published);
CREATE INDEX idx_brands_tenant_id_name ON brands(tenant_id, name);
CREATE INDEX idx_categories_tenant_id_parent_id ON categories(tenant_id, parent_id);
