-- Add missing created_at and updated_at columns to product_images table
ALTER TABLE product_images 
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Verify the table structure
\d product_images
