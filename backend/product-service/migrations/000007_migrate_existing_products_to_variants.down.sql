-- Migration: 000007_migrate_existing_products_to_variants (down)

-- Step 1: Clear default variant references
UPDATE products SET default_variant_id = NULL;

-- Step 2: Remove migrated variants
DELETE FROM product_variants; 