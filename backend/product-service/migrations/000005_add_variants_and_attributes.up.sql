-- Migration: 000005_add_variants_and_attributes

-- Step 1: Create the attributes table
CREATE TABLE attributes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL, -- e.g., 'Color', 'Size', 'Material'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_attributes_name ON attributes(name);
CREATE INDEX idx_attributes_deleted_at ON attributes(deleted_at);

-- Step 2: Create the product_variants table
CREATE TABLE product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    sku VARCHAR(100) UNIQUE NOT NULL, -- Unique SKU for each variant
    title VARCHAR(255), -- Optional: e.g., "Red - Large"
    price DECIMAL(10, 2) NOT NULL,
    discount_price DECIMAL(10, 2),
    inventory_qty INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_variant_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CONSTRAINT product_variants_price_check CHECK (price > 0),
    CONSTRAINT product_variants_inventory_qty_check CHECK (inventory_qty >= 0),
    CONSTRAINT product_variants_discount_price_check CHECK (discount_price IS NULL OR discount_price < price)
);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(sku);
CREATE INDEX idx_product_variants_deleted_at ON product_variants(deleted_at);

-- Step 3: Create the product_variant_attributes junction table
CREATE TABLE product_variant_attributes (
    product_variant_id UUID NOT NULL,
    attribute_id UUID NOT NULL,
    value VARCHAR(255) NOT NULL, -- e.g., 'Red', 'XL', 'Cotton'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (product_variant_id, attribute_id),
    CONSTRAINT fk_pva_variant FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    CONSTRAINT fk_pva_attribute FOREIGN KEY (attribute_id) REFERENCES attributes(id) ON DELETE CASCADE
);
CREATE INDEX idx_pva_variant_id ON product_variant_attributes(product_variant_id);
CREATE INDEX idx_pva_attribute_id ON product_variant_attributes(attribute_id);

-- Step 4: Modify the products table
-- Add a reference to a default variant (optional, but useful for listings/defaults)
ALTER TABLE products ADD COLUMN default_variant_id UUID NULL;
ALTER TABLE products ADD CONSTRAINT fk_product_default_variant FOREIGN KEY (default_variant_id) REFERENCES product_variants(id) ON DELETE SET NULL;

-- Remove variant-specific columns from the main products table
-- Note: This assumes data migration (copying existing product sku/price/qty to a default variant)
-- happens either before this step or via a separate script. For simplicity here, we just drop them.
ALTER TABLE products DROP COLUMN sku;
ALTER TABLE products DROP COLUMN price;
ALTER TABLE products DROP COLUMN discount_price;
ALTER TABLE products DROP COLUMN inventory_qty;

-- Optional: Add a check to ensure a product has at least one variant? (Could be complex)