@echo off
echo Running fix_product_data_api.go script...

REM Set the product ID to fix
set PRODUCT_ID=38c66092-140d-4488-a360-1b56f5affd42

REM Run the Go script
cd backend\product-service
go run scripts\fix_product_data_api.go %PRODUCT_ID%

echo Script execution completed.
