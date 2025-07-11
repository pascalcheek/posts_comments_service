FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk --no-cache add git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/internal/delivery/graphql/schema.graphql ./graphql/
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN addgroup -S appgroup && adduser -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./server", "-store=postgres", "-dsn=postgres://postgres:1234567890qwe@db:5432/posts_comments_db?sslmode=disable"]

