# Build Stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Rename the output binaries to avoid clashes with folder names
RUN go build -o goflash-api ./cmd/api
RUN go build -o goflash-worker ./cmd/worker

# Production Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/goflash-api .
COPY --from=builder /app/goflash-worker .

EXPOSE 8080
# Use the new name here
CMD ["./goflash-api"]