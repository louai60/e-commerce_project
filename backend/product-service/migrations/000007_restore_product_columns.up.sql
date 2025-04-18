-- Migration: 000007_restore_product_columns (Up)

-- Step 1: Add back columns to the products table
DO $$
BEGIN
    -- Add price column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'price') THEN
        ALTER TABLE products ADD COLUMN price DECIMAL(10, 2);
    END IF;

    -- Add discount_price column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'discount_price') THEN
        ALTER TABLE products ADD COLUMN discount_price DECIMAL(10, 2);
    END IF;

    -- Add sku column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'sku') THEN
        ALTER TABLE products ADD COLUMN sku VARCHAR(100);
    END IF;

    -- Add inventory_qty column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'inventory_qty') THEN
        ALTER TABLE products ADD COLUMN inventory_qty INT DEFAULT 0;
    END IF;

    -- Add inventory_status column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'inventory_status') THEN
        ALTER TABLE products ADD COLUMN inventory_status VARCHAR(50) DEFAULT 'in_stock';
    END IF;
END $$;

-- Step 2: Add back constraints (only if they don't exist)
DO $$
BEGIN
    -- Add price check constraint if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_price_check') THEN
        ALTER TABLE products ADD CONSTRAINT products_price_check CHECK (price > 0);
    END IF;

    -- Add inventory_qty check constraint if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_inventory_qty_check') THEN
        ALTER TABLE products ADD CONSTRAINT products_inventory_qty_check CHECK (inventory_qty >= 0);
    END IF;

    -- Add discount_price check constraint if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_discount_price_check') THEN
        ALTER TABLE products ADD CONSTRAINT products_discount_price_check CHECK (discount_price IS NULL OR discount_price < price);
    END IF;
END $$;
