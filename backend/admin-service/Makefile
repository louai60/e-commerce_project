.PHONY: proto build run test clean

# Generate protocol buffer code
proto:
    protoc --proto_path=. \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        proto/admin.proto

# Build the application
build: proto
    go build -o bin/admin-service main.go

# Run the service
run: build
    ./bin/admin-service

# Run tests
test:
    go test -v ./...

# Clean build artifacts
clean:
    rm -f proto/*.pb.go
    rm -rf bin/