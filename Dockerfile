# Build Stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install git for fetching private dependencies if needed
RUN apk add --no-cache git

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy EVERYTHING else
COPY . .

# IMPORTANT: Build the directory (package), not the specific file
# This is more robust against pathing quirks
RUN go build -o api ./cmd/api
RUN go build -o worker ./cmd/worker

# Production Stage
FROM alpine:latest
WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# Render uses the PORT environment variable
EXPOSE 8080

CMD ["./api"]