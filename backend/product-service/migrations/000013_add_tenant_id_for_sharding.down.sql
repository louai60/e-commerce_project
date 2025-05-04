-- Migration: 000012_add_tenant_id_for_sharding (Down)

-- Step 1: Drop composite indexes
DROP INDEX IF EXISTS idx_products_tenant_id_created_at;
DROP INDEX IF EXISTS idx_products_tenant_id_is_published;
DROP INDEX IF EXISTS idx_brands_tenant_id_name;
DROP INDEX IF EXISTS idx_categories_tenant_id_parent_id;

-- Step 2: Drop simple indexes
DROP INDEX IF EXISTS idx_products_tenant_id;
DROP INDEX IF EXISTS idx_brands_tenant_id;
DROP INDEX IF EXISTS idx_categories_tenant_id;

-- Step 3: Drop tenant_id columns
ALTER TABLE products DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE brands DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE categories DROP COLUMN IF EXISTS tenant_id;
