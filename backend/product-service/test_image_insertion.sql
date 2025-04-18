-- First, create a test product
INSERT INTO products (
    title, slug, description, short_description, price,
    sku, inventory_qty, inventory_status, is_published, created_at, updated_at
) VALUES (
    'Test Product for Image Test',
    'test-product-image-test',
    'This is a test product for image testing',
    'Test product for image testing',
    99.99,
    'TEST-SKU-' || extract(epoch from now())::text,
    100,
    'in_stock',
    true,
    now(),
    now()
) RETURNING id;

-- Now, insert a test image for this product
-- Replace 'product_id_here' with the ID returned from the previous query
INSERT INTO product_images (
    product_id, url, alt_text, position, created_at, updated_at
) VALUES (
    'product_id_here',  -- Replace this with the actual product ID
    'https://example.com/test-image.jpg',
    'Test Image',
    0,
    now(),
    now()
) RETURNING id;

-- Verify the image was inserted
SELECT * FROM product_images WHERE product_id = 'product_id_here';  -- Replace this with the actual product ID
