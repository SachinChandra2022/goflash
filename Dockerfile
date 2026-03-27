# Build Stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build with relative paths (Notice the ./ )
RUN go build -o api ./cmd/api/main.go
RUN go build -o worker ./cmd/worker/main.go

# Production Stage
FROM alpine:latest
WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# Render will use the PORT env var, but we'll expose 8081 locally
EXPOSE 8081

# Default command (overridden for worker in Render)
CMD ["./api"]