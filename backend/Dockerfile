# Stage 1: Build the binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile static binary with optimizations (strip debug symbols)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/main.go

# Stage 2: Final minimal image
FROM alpine:latest

WORKDIR /app

# Run as non-root user for security best practices
RUN adduser -D appuser
USER appuser

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/main .

EXPOSE 3000

CMD ["./main"]
