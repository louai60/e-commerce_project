.PHONY: build test proto docker clean

# Build the application
build:
    go build -o bin/user-service main.go

# Run tests
setup-test-db:
    psql -U postgres -c "DROP DATABASE IF EXISTS user_service_test;"
    psql -U postgres -c "CREATE DATABASE user_service_test;"

test: setup-test-db
    go test -v ./...

# Generate protocol buffer code
proto:
    protoc --proto_path=. \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        proto/user.proto

# Build docker image
docker:
    docker build -t user-service .

# Run the service
run:
    go run main.go

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f proto/*.pb.go

# Run integration tests
integration-test: setup-test-db
    go test -v ./tests/integration/...

# Generate mock objects
mocks:
    mockery --all --dir repository --output repository/mocks

# Run database migrations
migrate-up:
    migrate -path migrations -database "postgres://user:password@localhost:5432/nexcart_user?sslmode=disable" up

migrate-down:
    migrate -path migrations -database "postgres://user:password@localhost:5432/nexcart_user?sslmode=disable" down

# Add this new target
check-user:
    psql -U postgres -d nexcart_user -c "SELECT * FROM users WHERE email = 'alex@example.com';"

