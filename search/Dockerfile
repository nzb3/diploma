FROM golang:1.24.0-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target={$GOPATH}/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN go build -o server ./cmd/main/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/configs/ ./configs
EXPOSE 8080
CMD ["./server"]