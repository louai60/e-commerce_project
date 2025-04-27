-- Migration: 000010_restore_product_columns (Down)

-- Step 1: Remove the unique constraint on SKU
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_sku_unique;

-- Step 2: Remove the category_id column
ALTER TABLE products DROP COLUMN IF EXISTS category_id;

-- Step 3: Remove the inventory_status column
ALTER TABLE products DROP COLUMN IF EXISTS inventory_status;

-- Step 4: Remove the columns that were added back
ALTER TABLE products
DROP COLUMN IF EXISTS sku,
DROP COLUMN IF EXISTS price,
DROP COLUMN IF EXISTS discount_price,
DROP COLUMN IF EXISTS inventory_qty;
