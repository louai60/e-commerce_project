@echo off
echo Running fix_product_data.sql script...

REM Get database connection details from environment variables
set PGHOST=localhost
set PGPORT=5432
set PGDATABASE=product_service
set PGUSER=postgres
set PGPASSWORD=postgres

REM Run the SQL script
psql -h %PGHOST% -p %PGPORT% -d %PGDATABASE% -U %PGUSER% -f scripts/fix_product_data.sql

echo Script execution completed.
