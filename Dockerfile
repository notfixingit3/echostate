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

FROM alpine:3.20

# Web scans use browserless/chrome via BROWSER_WS_URL; traceroute for local path probes.
RUN apk add --no-cache ca-certificates traceroute

WORKDIR /app

COPY --from=builder /app/echostate /app/echostate

EXPOSE 8080

ENTRYPOINT ["/app/echostate"]
