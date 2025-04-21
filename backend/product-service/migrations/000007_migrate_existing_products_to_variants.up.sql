-- Migration: 000007_migrate_existing_products_to_variants

-- Step 1: Create a default variant for each existing product that doesn't have one
INSERT INTO product_variants (
    product_id,
    sku,
    title,
    price,
    discount_price,
    inventory_qty
)
SELECT 
    p.id as product_id,
    'SKU-' || p.id::text as sku,
    p.title,
    9.99 as price,
    NULL as discount_price,
    0 as inventory_qty
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id
WHERE p.deleted_at IS NULL
AND pv.id IS NULL;

-- Step 2: Update products table to set default_variant_id for products that don't have one
UPDATE products p
SET default_variant_id = pv.id
FROM product_variants pv
WHERE p.id = pv.product_id
AND p.deleted_at IS NULL
AND p.default_variant_id IS NULL; 