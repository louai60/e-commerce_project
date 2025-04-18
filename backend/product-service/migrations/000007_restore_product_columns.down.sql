-- Migration: 000007_restore_product_columns (Down)

-- Step 1: Remove constraints
DO $$
BEGIN
    -- Drop price check constraint if it exists
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_price_check') THEN
        ALTER TABLE products DROP CONSTRAINT products_price_check;
    END IF;

    -- Drop inventory_qty check constraint if it exists
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_inventory_qty_check') THEN
        ALTER TABLE products DROP CONSTRAINT products_inventory_qty_check;
    END IF;

    -- Drop discount_price check constraint if it exists
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_discount_price_check') THEN
        ALTER TABLE products DROP CONSTRAINT products_discount_price_check;
    END IF;
END $$;

-- Step 2: Remove columns
DO $$
BEGIN
    -- Drop price column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'price') THEN
        ALTER TABLE products DROP COLUMN price;
    END IF;

    -- Drop discount_price column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'discount_price') THEN
        ALTER TABLE products DROP COLUMN discount_price;
    END IF;

    -- Drop sku column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'sku') THEN
        ALTER TABLE products DROP COLUMN sku;
    END IF;

    -- Drop inventory_qty column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'inventory_qty') THEN
        ALTER TABLE products DROP COLUMN inventory_qty;
    END IF;

    -- Don't drop inventory_status as it might be used by other parts of the code
END $$;
