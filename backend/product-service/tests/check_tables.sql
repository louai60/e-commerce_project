-- Check if variant_images table exists
SELECT EXISTS (
    SELECT FROM information_schema.tables 
    WHERE table_name = 'variant_images'
);

-- List all tables in the database
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public'
ORDER BY table_name;

-- Check migration status
SELECT * FROM schema_migrations;
