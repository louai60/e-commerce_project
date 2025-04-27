@echo off
echo Running debug_product_data.sql script...

REM Get database connection details from environment variables
set PGHOST=localhost
set PGPORT=5432
set PGDATABASE=nexcart_product
set PGUSER=postgres
set PGPASSWORD=root

REM Run the SQL script and save output to a file
psql -h %PGHOST% -p %PGPORT% -d %PGDATABASE% -U %PGUSER% -f scripts/debug_product_data.sql > product_data_debug.txt

echo Script execution completed. Results saved to product_data_debug.txt
