@echo off

:: Start each service in a new terminal window
start cmd /k "cd api-gateway && air"
start cmd /k "cd product-service && air"
start cmd /k "cd user-service && air"

echo All services started with hot reloading:
echo API Gateway: http://localhost:8080
echo Product Service: localhost:50051
echo User Service: localhost:50052