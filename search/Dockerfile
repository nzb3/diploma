FROM golang:1.24.0-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/main/main.go

FROM alpine
WORKDIR /app
RUN apk add --no-cache libffi mupdf mupdf-dev
RUN find / -name "libmupdf.so*" -type f
RUN find / -name "libmupdf-dev.so*" -type f
COPY --from=builder /app/server .
COPY --from=builder /app/configs/ ./configs
EXPOSE 8080
CMD ["./server"]