-- Migration: 000008_add_composite_indexes (revert)

-- Drop all composite indexes created in the up migration

-- Drop product table composite indexes
DROP INDEX IF EXISTS idx_products_deleted_at_created_at;
DROP INDEX IF EXISTS idx_products_deleted_at_updated_at;
DROP INDEX IF EXISTS idx_products_category_id_deleted_at;
DROP INDEX IF EXISTS idx_products_brand_id_deleted_at;
DROP INDEX IF EXISTS idx_products_is_published_deleted_at;

-- Drop brand table composite indexes
DROP INDEX IF EXISTS idx_brands_deleted_at_created_at;

-- Drop category table composite indexes
DROP INDEX IF EXISTS idx_categories_deleted_at_created_at;
DROP INDEX IF EXISTS idx_categories_parent_id_deleted_at;
