@echo off

:: Create tmp directories if they don't exist
mkdir api-gateway\tmp 2>nul
mkdir product-service\tmp 2>nul
mkdir user-service\tmp 2>nul

:: Initial build for each service
cd api-gateway
go build -o ./tmp/main.exe .
cd ..

cd product-service
go build -o ./tmp/main.exe .
cd ..

cd user-service
go build -o ./tmp/main.exe .
cd ..

:: Start the services
call dev.bat