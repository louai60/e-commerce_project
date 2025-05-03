FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o inventory-service .

# Use a minimal alpine image for the final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/inventory-service .
COPY --from=builder /app/config ./config/

# Expose the port
EXPOSE 50055

# Command to run the executable
CMD ["./inventory-service"]
