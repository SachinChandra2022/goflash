# Build Stage
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build API and Worker binaries
RUN go build -o main-api cmd/api/main.go
RUN go build -o main-worker cmd/worker/main.go

# Run Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main-api .
COPY --from=builder /app/main-worker .
# Copy scripts for k6
COPY scripts/ ./scripts/ 

EXPOSE 8080
CMD ["./main-api"]