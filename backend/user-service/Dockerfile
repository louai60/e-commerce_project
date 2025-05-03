# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory for the build
WORKDIR /src

# Copy the entire backend directory to include all modules
# The context is set to ./backend in docker-compose.yml
COPY . .

# Set working directory to the user-service service
WORKDIR /src/user-service

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service .

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /src/user-service/user-service .

# Copy any necessary configuration files
COPY --from=builder /src/user-service/config ./config

# Expose the port
EXPOSE 50052

# Command to run the executable
CMD ["./user-service"]