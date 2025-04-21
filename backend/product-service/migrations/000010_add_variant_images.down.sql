-- Migration: 000008_add_variant_images (Down)

-- Step 1: Drop the variant_images table
DO $$ 
BEGIN
    -- Drop variant_images table if it exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'variant_images') THEN
        DROP TABLE variant_images;
    END IF;
END $$;
