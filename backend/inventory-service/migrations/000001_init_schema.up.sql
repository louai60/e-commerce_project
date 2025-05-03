-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create warehouses table
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create inventory_items table
CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    variant_id UUID,
    sku VARCHAR(100) NOT NULL UNIQUE,
    total_quantity INT NOT NULL DEFAULT 0,
    available_quantity INT NOT NULL DEFAULT 0,
    reserved_quantity INT NOT NULL DEFAULT 0,
    reorder_point INT NOT NULL DEFAULT 5,
    reorder_quantity INT NOT NULL DEFAULT 20,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_STOCK',
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT inventory_quantities_check CHECK (available_quantity >= 0 AND reserved_quantity >= 0 AND total_quantity >= 0)
);

-- Create inventory_locations table
CREATE TABLE inventory_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id UUID NOT NULL,
    warehouse_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    available_quantity INT NOT NULL DEFAULT 0,
    reserved_quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_inventory_item FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    CONSTRAINT inventory_location_unique UNIQUE (inventory_item_id, warehouse_id),
    CONSTRAINT inventory_location_quantities_check CHECK (available_quantity >= 0 AND reserved_quantity >= 0 AND quantity >= 0)
);

-- Create inventory_transactions table
CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id UUID NOT NULL,
    warehouse_id UUID,
    transaction_type VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50),
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_inventory_item_transaction FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_warehouse_transaction FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE SET NULL
);

-- Create inventory_reservations table
CREATE TABLE inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_item_id UUID NOT NULL,
    warehouse_id UUID,
    quantity INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    expiration_time TIMESTAMPTZ NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_inventory_item_reservation FOREIGN KEY (inventory_item_id) REFERENCES inventory_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_warehouse_reservation FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE SET NULL,
    CONSTRAINT inventory_reservation_quantity_check CHECK (quantity > 0)
);

-- Create indexes for performance
CREATE INDEX idx_inventory_items_product_id ON inventory_items(product_id);
CREATE INDEX idx_inventory_items_variant_id ON inventory_items(variant_id);
CREATE INDEX idx_inventory_items_sku ON inventory_items(sku);
CREATE INDEX idx_inventory_items_status ON inventory_items(status);
CREATE INDEX idx_inventory_locations_warehouse_id ON inventory_locations(warehouse_id);
CREATE INDEX idx_inventory_transactions_inventory_item_id ON inventory_transactions(inventory_item_id);
CREATE INDEX idx_inventory_transactions_reference_id ON inventory_transactions(reference_id);
CREATE INDEX idx_inventory_reservations_inventory_item_id ON inventory_reservations(inventory_item_id);
CREATE INDEX idx_inventory_reservations_status ON inventory_reservations(status);
CREATE INDEX idx_inventory_reservations_expiration_time ON inventory_reservations(expiration_time);
