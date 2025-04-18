-- Revert Step 5: Drop partial index
DROP INDEX IF EXISTS idx_products_published_not_deleted;

-- Revert Step 4: Drop indexes on deleted_at columns
DROP INDEX IF EXISTS idx_brands_deleted_at;
DROP INDEX IF EXISTS idx_categories_deleted_at;
-- Keep idx_products_deleted_at if it was created by migration 000003

-- Revert Step 3: Drop CHECK constraints
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_price_check;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_inventory_qty_check;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_discount_price_check;

-- Revert Step 2: Remove deleted_at columns
ALTER TABLE brands DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE categories DROP COLUMN IF EXISTS deleted_at;

-- Revert Step 1: Revert timestamp columns back to TIMESTAMP (potential data loss if TZ was used)
-- Note: Reverting TIMESTAMPTZ to TIMESTAMP might lose timezone information.
-- This is a best effort rollback.
ALTER TABLE products ALTER COLUMN created_at TYPE TIMESTAMP;
ALTER TABLE products ALTER COLUMN updated_at TYPE TIMESTAMP;

ALTER TABLE brands ALTER COLUMN created_at TYPE TIMESTAMP;
ALTER TABLE brands ALTER COLUMN updated_at TYPE TIMESTAMP;

ALTER TABLE categories ALTER COLUMN created_at TYPE TIMESTAMP;
ALTER TABLE categories ALTER COLUMN updated_at TYPE TIMESTAMP;

ALTER TABLE product_images ALTER COLUMN created_at TYPE TIMESTAMP;
ALTER TABLE product_images ALTER COLUMN updated_at TYPE TIMESTAMP;
