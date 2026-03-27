FROM golang:1.24-alpine AS builder 
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api cmd/api/main.go
RUN go build -o worker cmd/worker/main.go

# Production Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# Render defaults to port 10000 if not specified, 
# but we will use 8080 or 8081 as per your env.
EXPOSE 8081 

CMD ["./api"]