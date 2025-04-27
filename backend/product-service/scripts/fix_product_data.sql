-- Fix product data script
-- This script updates product data to ensure all fields are properly populated

-- 1. Update product images to use the correct URLs from the POST request
UPDATE product_images
SET url = 'https://example.com/images/watch-main.jpg',
    alt_text = 'Smart Fitness Tracker Watch on wrist'
WHERE product_id = '38c66092-140d-4488-a360-1b56f5affd42' AND position = 1;

INSERT INTO product_images (id, product_id, url, alt_text, position, created_at, updated_at)
VALUES (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'https://example.com/images/watch-side.jpg', 'Side view showing touchscreen interface', 2, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 2. Update product price
UPDATE products
SET price = 199.99,
    discount_price = 179.99,
    sku = 'WATCH-001',
    inventory_qty = 200
WHERE id = '38c66092-140d-4488-a360-1b56f5affd42';

-- 3. Update product specifications
INSERT INTO product_specifications (id, product_id, name, value, unit, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'Battery Life', '7', 'days', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'Compatibility', 'iOS & Android', NULL, NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'Display', '1.78', 'inch AMOLED', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'Sensors', 'Optical HR, GPS, SpO2', NULL, NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'Warranty', '2', 'years', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'Water Resistance', '5 ATM', NULL, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 4. Update product tags
INSERT INTO product_tags (id, product_id, tag, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'fitness tracker', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'GPS', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'health', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'smartwatch', NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'wearable', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 5. Update product inventory locations
INSERT INTO product_inventory_locations (id, product_id, warehouse_id, quantity, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'A1', 100, NOW(), NOW()),
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'B2', 100, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 6. Update product shipping
INSERT INTO product_shipping (id, product_id, free_shipping, estimated_days, express_shipping_available, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', true, '0', false, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 7. Update product SEO
UPDATE product_seo
SET meta_title = 'Smart Fitness Tracker Watch | Health Monitoring | Your Brand',
    meta_description = 'Track workouts, monitor health metrics, and stay connected with our advanced waterproof smartwatch featuring GPS and 7-day battery life.'
WHERE product_id = '38c66092-140d-4488-a360-1b56f5affd42';

-- 8. Update product discounts
INSERT INTO product_discounts (id, product_id, type, value, expires_at, created_at, updated_at)
VALUES 
    (gen_random_uuid(), '38c66092-140d-4488-a360-1b56f5affd42', 'percentage', 10, NOW() + INTERVAL '1 year', NOW(), NOW())
ON CONFLICT DO NOTHING;
