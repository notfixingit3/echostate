# Stage 1: Package pre-built Linux binaries (CI multi-arch — avoids QEMU Go compile on arm64)
FROM alpine:3.20 AS production-prebuilt

# Web scans use browserless/chrome via BROWSER_WS_URL; traceroute for local path probes.
RUN apk add --no-cache ca-certificates traceroute

WORKDIR /app

ARG TARGETARCH
COPY bin/echostate-linux-${TARGETARCH} /app/echostate
RUN chmod +x /app/echostate

EXPOSE 8080
ENTRYPOINT ["/app/echostate"]

# Stage 2: Build inside Docker (local single-arch / docker-compose.dev)
FROM golang:1.26-alpine AS builder

WORKDIR /app

ARG VERSION=dev
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/notfixingit3/echostate/internal/version.Version=${VERSION}" \
    -o echostate ./main.go

FROM alpine:3.20 AS production

# Web scans use browserless/chrome via BROWSER_WS_URL; traceroute for local path probes.
RUN apk add --no-cache ca-certificates traceroute

WORKDIR /app

COPY --from=builder /app/echostate /app/echostate

EXPOSE 8080
ENTRYPOINT ["/app/echostate"]