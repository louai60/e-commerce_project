-- Migration: 000013_remove_inventory_fields (Up)

-- Step 1: Remove inventory-related columns from products table
ALTER TABLE products
DROP COLUMN IF EXISTS inventory_qty,
DROP COLUMN IF EXISTS inventory_status;

-- Step 2: Remove inventory-related columns from product_variants table
ALTER TABLE product_variants
DROP COLUMN IF EXISTS inventory_qty;

-- Step 3: Drop the product_inventory_locations table
DROP TABLE IF EXISTS product_inventory_locations CASCADE;

-- Step 4: Keep SKU in both services for now as a reference key
-- This allows for easier integration during the transition period
-- In a future migration, we could consider removing SKU from products
-- once the inventory service is fully established

-- Step 5: Add a comment to document the change
COMMENT ON TABLE products IS 'Core product information. Inventory data has been moved to the inventory service.';
