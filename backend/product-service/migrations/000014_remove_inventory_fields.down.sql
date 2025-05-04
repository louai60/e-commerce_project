-- Migration: 000013_remove_inventory_fields (Down)

-- Step 1: Add back inventory-related columns to products table
ALTER TABLE products
ADD COLUMN IF NOT EXISTS inventory_qty INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS inventory_status VARCHAR(20) DEFAULT 'out_of_stock';

-- Step 2: Add back inventory-related columns to product_variants table
ALTER TABLE product_variants
ADD COLUMN IF NOT EXISTS inventory_qty INTEGER DEFAULT 0;

-- Step 3: Recreate the product_inventory_locations table
CREATE TABLE IF NOT EXISTS product_inventory_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    warehouse_id VARCHAR(100) NOT NULL,
    available_qty INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_product_inventory_location FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Step 4: Remove the comment
COMMENT ON TABLE products IS NULL;
