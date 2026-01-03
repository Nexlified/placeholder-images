# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o avata-go ./cmd/avata-go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/avata-go .

# Expose port
EXPOSE 8080

# Set default environment variables
ENV ADDR=":8080"
ENV CACHE_SIZE="2000"

# Run the application
CMD ["./avata-go"]

