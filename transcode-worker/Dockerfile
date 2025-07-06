# -------- Builder Stage --------
    FROM golang:1.22 AS builder

    WORKDIR /app
    
    # Copy go.mod and go.sum first to leverage Docker layer caching
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the source code
    COPY . .
    
    # Build the Go binary
    RUN CGO_ENABLED=0 GOOS=linux go build -o transcode-worker
    
    # -------- Final Stage --------
    FROM debian:bullseye-slim
    
    # Install FFmpeg
    RUN apt-get update && \
        apt-get install -y ffmpeg ca-certificates && \
        apt-get clean && \
        rm -rf /var/lib/apt/lists/*
    
    # Create app directory
    WORKDIR /app
    
    # Copy the binary from builder
    COPY --from=builder /app/transcode-worker .
    
    # Run the service
    ENTRYPOINT ["./transcode-worker"]
    