# Dockerfile
FROM golang:1.23-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# watcher-server
FROM base AS watcher-server
COPY . .
RUN go build -o watcher-server ./cmd/server
CMD ["./watcher-server"]

# handler-client
FROM base AS handler-client
COPY . .
RUN go build -o handler-client ./cmd/client
CMD ["./handler-client"]
