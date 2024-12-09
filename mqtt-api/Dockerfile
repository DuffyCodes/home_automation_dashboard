FROM golang:1.20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy Go module files for mqtt-api and shared module
COPY mqtt-api/go.mod mqtt-api/go.sum ./mqtt-api/
COPY shared/go.mod shared/go.sum ./shared/

COPY shared/ ./shared

# Download dependencies
WORKDIR /app/mqtt-api
RUN go mod tidy

# Copy the source code for mqtt-api and shared module
COPY mqtt-api/ .

# Build the mqtt-api binary
WORKDIR /app/mqtt-api
RUN go build -o mqtt-api ./cmd

# Use a minimal image for the runtime
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/mqtt-api/mqtt-api .  

# Expose the port used by mqtt-api
EXPOSE 8080

# Set the entry point
CMD ["./mqtt-api"]