-- Migration: 000006_add_product_extensions (Down)

-- Step 1: Remove inventory_status from products table
ALTER TABLE products DROP COLUMN IF EXISTS inventory_status;

-- Step 2: Drop product_inventory_locations table
DROP TABLE IF EXISTS product_inventory_locations CASCADE;

-- Step 3: Drop product_discounts table
DROP TABLE IF EXISTS product_discounts CASCADE;

-- Step 4: Drop product_shipping table
DROP TABLE IF EXISTS product_shipping CASCADE;

-- Step 5: Drop product_seo table
DROP TABLE IF EXISTS product_seo CASCADE;

-- Step 6: Drop product_attributes table
DROP TABLE IF EXISTS product_attributes CASCADE;

-- Step 7: Drop product_specifications table
DROP TABLE IF EXISTS product_specifications CASCADE;

-- Step 8: Drop product_tags table
DROP TABLE IF EXISTS product_tags CASCADE;
