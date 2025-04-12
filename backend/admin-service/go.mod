module github.com/louai60/e-commerce_project/backend/admin-service

go 1.24.0

require (
	github.com/joho/godotenv v1.5.1
	github.com/louai60/e-commerce_project/backend/product-service v0.0.0-00010101000000-000000000000
	github.com/louai60/e-commerce_project/backend/user-service v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.5
)

require (
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
)

replace github.com/louai60/e-commerce_project/backend/product-service => ../product-service

replace github.com/louai60/e-commerce_project/backend/user-service => ../user-service
