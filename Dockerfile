# --- Stage 1: Build ---
FROM golang:1.24.5 AS builder

WORKDIR /app
COPY . .

# Ensure dependencies are tidy and compile test server
WORKDIR /app/cmd/simulator-core-api-test-server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/test-server

# --- Stage 2: Minimal runtime ---
FROM alpine:3.20

# Add CA certs if needed (e.g., for HTTPS)
RUN apk add --no-cache ca-certificates

COPY --from=builder /out/test-server /simulator-core-api-test-server
ENTRYPOINT ["/simulator-core-api-test-server"]