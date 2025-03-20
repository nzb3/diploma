FROM golang:1.24.0-alpine AS builder
WORKDIR /app
COPY .. .
RUN go mod download
RUN go build -o server .

FROM alpine
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]