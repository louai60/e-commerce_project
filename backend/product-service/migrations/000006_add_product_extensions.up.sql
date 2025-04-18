-- Migration: 000006_add_product_extensions

-- Step 1: Create product_tags table
CREATE TABLE product_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    tag VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_tag_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_product_tags_product_id ON product_tags(product_id);
CREATE UNIQUE INDEX idx_product_tags_product_tag ON product_tags(product_id, tag);

-- Step 2: Create product_specifications table
CREATE TABLE product_specifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL,
    unit VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_spec_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_product_specs_product_id ON product_specifications(product_id);

-- Step 3: Create product_attributes table (for product-level attributes)
CREATE TABLE product_attributes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_attr_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_product_attributes_product_id ON product_attributes(product_id);
CREATE UNIQUE INDEX idx_product_attributes_product_name ON product_attributes(product_id, name);

-- Step 4: Create product_seo table
CREATE TABLE product_seo (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL UNIQUE,
    meta_title VARCHAR(255),
    meta_description TEXT,
    keywords TEXT[], -- Array of keywords
    tags TEXT[], -- Array of SEO tags
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_seo_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_product_seo_product_id ON product_seo(product_id);

-- Step 5: Create product_shipping table
CREATE TABLE product_shipping (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL UNIQUE,
    free_shipping BOOLEAN DEFAULT false,
    estimated_days INTEGER,
    express_available BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_shipping_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_product_shipping_product_id ON product_shipping(product_id);

-- Step 6: Create product_discounts table
CREATE TABLE product_discounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    discount_type VARCHAR(50) NOT NULL, -- 'percentage', 'fixed', etc.
    value DECIMAL(10,2) NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_discount_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_product_discounts_product_id ON product_discounts(product_id);
CREATE INDEX idx_product_discounts_expires_at ON product_discounts(expires_at);

-- Step 7: Create product_inventory_locations table
CREATE TABLE product_inventory_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    warehouse_id VARCHAR(100) NOT NULL,
    available_qty INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_product_inventory_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CONSTRAINT product_inventory_available_qty_check CHECK (available_qty >= 0)
);
CREATE INDEX idx_product_inventory_product_id ON product_inventory_locations(product_id);
CREATE UNIQUE INDEX idx_product_inventory_product_warehouse ON product_inventory_locations(product_id, warehouse_id);

-- Step 8: Add inventory_status to products table
ALTER TABLE products ADD COLUMN inventory_status VARCHAR(50) DEFAULT 'in_stock';
