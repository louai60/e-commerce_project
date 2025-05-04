# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory for the build
WORKDIR /src

# Copy the entire backend directory to include all modules
# The context is set to ./backend in docker-compose.yml
COPY . .

# Set working directory to the api-gateway service
WORKDIR /src/api-gateway

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /src/api-gateway/api-gateway .

# Copy any necessary configuration files
# Note: We don't copy .env.example to .env as environment variables will be provided by docker-compose
COPY --from=builder /src/api-gateway/config ./config

# Create uploads directory
RUN mkdir -p ./uploads

# Expose the port
EXPOSE 8080

# Command to run the executable
CMD ["./api-gateway"]
