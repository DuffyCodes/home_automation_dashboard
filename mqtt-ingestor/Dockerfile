FROM golang:1.20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy Go module files
COPY mqtt-ingestor/go.mod mqtt-ingestor/go.sum ./mqtt-ingestor/
COPY shared/go.mod shared/go.sum ./shared/

COPY shared/ ./shared

# Download dependencies
WORKDIR /app/mqtt-ingestor
RUN go mod tidy

# Copy the source code for mqtt-ingestor and shared module
COPY mqtt-ingestor/ .

# Build the mqtt-ingestor binary
WORKDIR /app/mqtt-ingestor
RUN go build -o mqtt-ingestor ./cmd

# Use a minimal image for the runtime
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/mqtt-ingestor/mqtt-ingestor .  

# Expose the port used by mqtt-ingestor
EXPOSE 8080

# Set the entry point
CMD ["./mqtt-ingestor"]
