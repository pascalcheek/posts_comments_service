FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go install github.com/99designs/gqlgen@v0.17.24
RUN go run github.com/99designs/gqlgen generate

RUN go build -o server ./cmd/server

EXPOSE 8080
CMD ["./server"]
