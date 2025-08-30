# syntax=docker/dockerfile:1

### Build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (statically linked)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

### Run stage (minimal runtime)
FROM alpine:3.20 AS production

# Set working directory
WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
ENV PORT=8080
EXPOSE 8080

# Run the binary
CMD ["./main"]
