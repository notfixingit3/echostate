FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o echostate ./main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates chromium

WORKDIR /app

COPY --from=builder /app/echostate /app/echostate

EXPOSE 8080

ENTRYPOINT ["/app/echostate"]
