FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api  ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/etl  ./cmd/etl

# ── API image ──────────────────────────────────────────────────────────────────
FROM alpine:3.19 AS api
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]

# ── ETL image ──────────────────────────────────────────────────────────────────
FROM alpine:3.19 AS etl
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/etl /app/etl
ENTRYPOINT ["/app/etl"]