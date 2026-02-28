## Multi-stage Dockerfile for Go Gin backend on Back4App

FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build deps (e.g. git) if needed
RUN apk add --no-cache ca-certificates git

# Pre-copy go.mod/go.sum to leverage Docker layer cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .


# ---- Runtime image ----
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

# Copy binary from builder
COPY --from=builder /app/app /app/app

# Environment (Back4App will set PORT and DATABASE_URL)
ENV GIN_MODE=release

# Default port your app listens on (matches main.go default)
EXPOSE 8080

CMD ["/app/app"]

