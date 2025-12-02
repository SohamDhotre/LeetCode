# Multi-stage build for smaller final image
FROM golang:1.21-alpine AS builder

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source code (needed for go mod tidy)
COPY sync.go .

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sync .

# Final stage - minimal runtime image
FROM alpine:latest

# Install git and timezone data
RUN apk --no-cache add git ca-certificates tzdata

# Create app directory
WORKDIR /workspace

# Copy the binary from builder
COPY --from=builder /app/sync /usr/local/bin/sync

# Copy configuration files
COPY config.json .

# Set the entrypoint
ENTRYPOINT ["sync"]
