.PHONY: test test-product test-user test-api-gateway

# Run all tests
test: test-product test-user test-api-gateway

# Run product service tests
test-product:
	@echo "Running product service tests..."
	cd product-service && go test ./...

# Run user service tests
test-user:
	@echo "Running user service tests..."
	cd user-service && go test ./...

# Run API gateway tests
test-api-gateway:
	@echo "Running API gateway tests..."
	cd api-gateway && go test ./...
