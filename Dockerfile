# STEP 1: Build the binaries
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o api cmd/api/main.go
RUN go build -o worker cmd/worker/main.go

# STEP 2: Create the production image
FROM alpine:latest
WORKDIR /root/
# Copy the compiled binaries from the builder
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
# Copy the .env file if you have one (optional, Render handles env vars)
# EXPOSE the port
EXPOSE 8080
# We will override the CMD in the Render dashboard
CMD ["./api"]