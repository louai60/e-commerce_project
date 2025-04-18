-- Migration: 000005_add_variants_and_attributes (Down)

-- Step 1: Add back columns to the products table
-- Note: Data restoration from variants would need a separate script/manual process.
ALTER TABLE products ADD COLUMN sku VARCHAR(100) UNIQUE;
ALTER TABLE products ADD COLUMN price DECIMAL(10, 2);
ALTER TABLE products ADD COLUMN discount_price DECIMAL(10, 2);
ALTER TABLE products ADD COLUMN inventory_qty INT DEFAULT 0;

-- Add back constraints if they were specific to these columns (check previous migrations if needed)
-- Example (adjust based on actual constraints removed implicitly or explicitly):
-- ALTER TABLE products ADD CONSTRAINT products_price_check CHECK (price > 0);
-- ALTER TABLE products ADD CONSTRAINT products_inventory_qty_check CHECK (inventory_qty >= 0);
-- ALTER TABLE products ADD CONSTRAINT products_discount_price_check CHECK (discount_price IS NULL OR discount_price < price);

-- Step 2: Remove the default_variant_id column and constraint
ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_product_default_variant;
ALTER TABLE products DROP COLUMN IF EXISTS default_variant_id;

-- Step 3: Drop the junction table
DROP TABLE IF EXISTS product_variant_attributes;

-- Step 4: Drop the product_variants table
DROP TABLE IF EXISTS product_variants;

-- Step 5: Drop the attributes table
DROP TABLE IF EXISTS attributes;