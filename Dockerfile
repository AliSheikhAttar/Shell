# Dockerfile for the Go application

# Stage 1: Builder stage
FROM golang:latest AS builder

WORKDIR /app

# Copy go mod and sum files to download dependencies first
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the entire project source code
COPY . .

# Build the Go application - adjust the entrypoint path if needed
RUN go build -o my-go-app ./cmd/main.go

# Stage 2: Final stage - Create a minimal image
FROM alpine:latest

WORKDIR /app

# Copy necessary files from the builder stage
COPY --from=builder /app/my-go-app .
# If you have any static files or other dependencies, copy them here as well

# Command to run your Go application
CMD ["./my-go-app"]