-- Drop indexes
DROP INDEX IF EXISTS idx_inventory_reservations_expiration_time;
DROP INDEX IF EXISTS idx_inventory_reservations_status;
DROP INDEX IF EXISTS idx_inventory_reservations_inventory_item_id;
DROP INDEX IF EXISTS idx_inventory_transactions_reference_id;
DROP INDEX IF EXISTS idx_inventory_transactions_inventory_item_id;
DROP INDEX IF EXISTS idx_inventory_locations_warehouse_id;
DROP INDEX IF EXISTS idx_inventory_items_status;
DROP INDEX IF EXISTS idx_inventory_items_sku;
DROP INDEX IF EXISTS idx_inventory_items_variant_id;
DROP INDEX IF EXISTS idx_inventory_items_product_id;

-- Drop tables
DROP TABLE IF EXISTS inventory_reservations;
DROP TABLE IF EXISTS inventory_transactions;
DROP TABLE IF EXISTS inventory_locations;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS warehouses;
