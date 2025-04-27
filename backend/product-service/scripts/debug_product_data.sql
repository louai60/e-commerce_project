-- Debug product data script
-- This script queries all product data from the database for a specific product ID

-- Set the product ID to query
\set product_id '38c66092-140d-4488-a360-1b56f5affd42'

-- 1. Query the main product data
SELECT * FROM products WHERE id = :'product_id';

-- 2. Query product images
SELECT * FROM product_images WHERE product_id = :'product_id';

-- 3. Query product specifications
SELECT * FROM product_specifications WHERE product_id = :'product_id';

-- 4. Query product tags
SELECT * FROM product_tags WHERE product_id = :'product_id';

-- 5. Query product categories
SELECT pc.*, c.name as category_name 
FROM product_categories pc
JOIN categories c ON pc.category_id = c.id
WHERE pc.product_id = :'product_id';

-- 6. Query product inventory locations
SELECT * FROM product_inventory_locations WHERE product_id = :'product_id';

-- 7. Query product shipping
SELECT * FROM product_shipping WHERE product_id = :'product_id';

-- 8. Query product SEO
SELECT * FROM product_seo WHERE product_id = :'product_id';

-- 9. Query product discounts
SELECT * FROM product_discounts WHERE product_id = :'product_id';

-- 10. Query product variants
SELECT * FROM product_variants WHERE product_id = :'product_id';

-- 11. Query product attributes
SELECT * FROM product_attributes WHERE product_id = :'product_id';
