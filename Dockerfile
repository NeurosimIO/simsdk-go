FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/simulator-core-api-test-server ./cmd/simulator-core-api-test-server
WORKDIR /app/cmd/simulator-core-api-test-server
RUN go build -o /simulator-core-api-test-server

FROM alpine:3.20
COPY --from=builder /simulator-core-api-test-server /simulator-core-api-test-server
ENTRYPOINT ["/simulator-core-api-test-server"]