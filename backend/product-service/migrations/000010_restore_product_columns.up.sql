-- Migration: 000010_restore_product_columns (Up)

-- Step 1: Add back the columns that were removed in previous migrations
ALTER TABLE products
ADD COLUMN IF NOT EXISTS sku VARCHAR(100),
ADD COLUMN IF NOT EXISTS price DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS discount_price DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS inventory_qty INTEGER;

-- Step 2: Update the products table with data from the default variants
UPDATE products p
SET 
    sku = v.sku,
    price = v.price,
    discount_price = v.discount_price,
    inventory_qty = v.inventory_qty
FROM product_variants v
WHERE v.product_id = p.id AND v.is_default = true;

-- Step 3: Set default values for products without default variants
UPDATE products
SET 
    sku = CONCAT('PROD-', SUBSTRING(id::text, 1, 8)),
    price = 0,
    inventory_qty = 0
WHERE sku IS NULL;

-- Step 4: Add NOT NULL constraints to required columns
ALTER TABLE products
ALTER COLUMN sku SET NOT NULL,
ALTER COLUMN price SET NOT NULL,
ALTER COLUMN inventory_qty SET NOT NULL;

-- Step 5: Add unique constraint to SKU
ALTER TABLE products
ADD CONSTRAINT products_sku_unique UNIQUE (sku);

-- Step 6: Add inventory_status column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'products' AND column_name = 'inventory_status'
    ) THEN
        ALTER TABLE products ADD COLUMN inventory_status VARCHAR(20) DEFAULT 'in_stock';
    END IF;
END $$;

-- Step 7: Update inventory_status based on inventory_qty
UPDATE products
SET inventory_status = 
    CASE 
        WHEN inventory_qty > 0 THEN 'in_stock'
        WHEN inventory_qty = 0 THEN 'out_of_stock'
        ELSE 'backorder'
    END;

-- Step 8: Add category_id column for direct category reference
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'products' AND column_name = 'category_id'
    ) THEN
        ALTER TABLE products ADD COLUMN category_id UUID REFERENCES categories(id);
        
        -- Update category_id from product_categories for primary category
        UPDATE products p
        SET category_id = pc.category_id
        FROM product_categories pc
        WHERE p.id = pc.product_id;
    END IF;
END $$;
