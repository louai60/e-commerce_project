-- Step 1: Standardize timestamp columns to TIMESTAMPTZ
ALTER TABLE products ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
ALTER TABLE products ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE brands ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
ALTER TABLE brands ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE categories ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
ALTER TABLE categories ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE product_images ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
ALTER TABLE product_images ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

-- Step 2: Add deleted_at columns for soft deletes (products already has it from migration 000003)
ALTER TABLE brands ADD COLUMN deleted_at TIMESTAMPTZ NULL;
ALTER TABLE categories ADD COLUMN deleted_at TIMESTAMPTZ NULL;

-- Step 3: Add CHECK constraints for data integrity
ALTER TABLE products ADD CONSTRAINT products_price_check CHECK (price > 0);
ALTER TABLE products ADD CONSTRAINT products_inventory_qty_check CHECK (inventory_qty >= 0);
-- Allow discount_price to be NULL or less than price
ALTER TABLE products ADD CONSTRAINT products_discount_price_check CHECK (discount_price IS NULL OR discount_price < price);

-- Step 4: Add indexes on deleted_at columns
CREATE INDEX idx_brands_deleted_at ON brands(deleted_at);
CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);
-- Index on products.deleted_at might already exist from migration 000003, but ensure it's there
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);


-- Step 5: Add partial index for performance on common queries
CREATE INDEX idx_products_published_not_deleted ON products(is_published) WHERE deleted_at IS NULL AND is_published = true;
