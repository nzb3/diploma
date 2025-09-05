# Debug build stage
FROM golang:1.24.0-alpine AS builder-debug
WORKDIR /app
RUN apk add --no-cache build-base musl-dev
RUN apk add mupdf-dev=1.24.10-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.21/community
RUN go install github.com/go-delve/delve/cmd/dlv@latest
ENV CGO_ENABLED=1
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -tags musl -gcflags="all=-N -l" -o server ./cmd/main/main.go

FROM alpine AS debug
WORKDIR /app
RUN apk add mupdf-dev=1.24.10-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.21/community
COPY --from=builder-debug /app/server .
COPY --from=builder-debug /go/bin/dlv .
COPY --from=builder-debug /app/configs/config.yml .
EXPOSE 8082 40001
CMD ["./dlv", "--listen=:40001", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./server"]

# Release build stage
FROM golang:1.24.0-alpine AS builder-release
WORKDIR /app
RUN apk add --no-cache build-base musl-dev
RUN apk add mupdf-dev=1.24.10-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.21/community
ENV CGO_ENABLED=1
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -tags musl -o server ./cmd/main/main.go

FROM alpine AS release
ENV GIN_MODE=release
WORKDIR /app
RUN apk add mupdf-dev=1.24.10-r0 --repository=https://dl-cdn.alpinelinux.org/alpine/v3.21/community
COPY --from=builder-release /app/server .
COPY --from=builder-release /app/configs/config.yml .
EXPOSE 8082
CMD ["./server"]
