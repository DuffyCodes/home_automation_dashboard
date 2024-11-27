FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o mqtt-ingestor ./cmd

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/mqtt-ingestor .
EXPOSE 8080
CMD ["./mqtt-ingestor"]
