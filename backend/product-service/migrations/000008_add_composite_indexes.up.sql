-- Migration: 000008_add_composite_indexes

-- Step 1: Add composite indexes for soft deletes to improve query performance
-- These indexes will significantly speed up queries that filter by deleted_at IS NULL
-- and sort by created_at or other columns

-- Composite index for products table: (deleted_at, created_at)
-- This will optimize queries that filter for non-deleted products and sort by creation date
CREATE INDEX IF NOT EXISTS idx_products_deleted_at_created_at ON products(deleted_at, created_at);

-- Composite index for products table: (deleted_at, updated_at)
-- This will optimize queries that filter for non-deleted products and sort by update date
CREATE INDEX IF NOT EXISTS idx_products_deleted_at_updated_at ON products(deleted_at, updated_at);

-- Composite index for products table: (category_id, deleted_at)
-- This will optimize queries that filter products by category and deletion status
CREATE INDEX IF NOT EXISTS idx_products_category_id_deleted_at ON products(category_id, deleted_at);

-- Composite index for products table: (brand_id, deleted_at)
-- This will optimize queries that filter products by brand and deletion status
CREATE INDEX IF NOT EXISTS idx_products_brand_id_deleted_at ON products(brand_id, deleted_at);

-- Composite index for products table: (is_published, deleted_at)
-- This will optimize queries that filter products by publication status and deletion status
CREATE INDEX IF NOT EXISTS idx_products_is_published_deleted_at ON products(is_published, deleted_at);

-- Composite index for brands table: (deleted_at, created_at)
-- This will optimize queries that filter for non-deleted brands and sort by creation date
CREATE INDEX IF NOT EXISTS idx_brands_deleted_at_created_at ON brands(deleted_at, created_at);

-- Composite index for categories table: (deleted_at, created_at)
-- This will optimize queries that filter for non-deleted categories and sort by creation date
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at_created_at ON categories(deleted_at, created_at);

-- Composite index for categories table: (parent_id, deleted_at)
-- This will optimize queries that filter categories by parent and deletion status
CREATE INDEX IF NOT EXISTS idx_categories_parent_id_deleted_at ON categories(parent_id, deleted_at);
