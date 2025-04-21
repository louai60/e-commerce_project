-- Migration: 000008_add_variant_images (Up)

-- Step 1: Create the variant_images table
DO $$ 
BEGIN
    -- Create variant_images table if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'variant_images') THEN
        CREATE TABLE variant_images (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            variant_id UUID NOT NULL,
            url TEXT NOT NULL,
            alt_text TEXT,
            position INT DEFAULT 0,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            CONSTRAINT fk_variant_image FOREIGN KEY (variant_id) REFERENCES product_variants(id) ON DELETE CASCADE
        );
        
        -- Create index on variant_id for faster lookups
        CREATE INDEX idx_variant_images_variant_id ON variant_images(variant_id);
    END IF;
END $$;
