FROM golang:1.24.0-alpine AS builder-debug
WORKDIR /app
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -gcflags="all=-N -l" -o server ./cmd/main/main.go

FROM alpine AS debug
WORKDIR /app
RUN apk add mupdf-dev=1.24.10-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.21/community
COPY --from=builder-debug /app/server .
COPY --from=builder-debug /app/configs/ ./configs
COPY --from=builder-debug /go/bin/dlv .
EXPOSE 8081 40000
CMD ["./dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./server"]

# Release build stage
FROM golang:1.24.0-alpine AS builder-release
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/main/main.go

FROM alpine AS release
ENV GIN_MODE=release
WORKDIR /app
RUN apk add mupdf-dev=1.24.10-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.21/community
COPY --from=builder-release /app/server .
COPY --from=builder-release /app/configs/ ./configs
EXPOSE 8081
CMD ["./server"]
