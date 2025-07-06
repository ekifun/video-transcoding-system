# ---------- Builder ----------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    # Fetch Redis v9 and tidy dependencies
    RUN go get github.com/redis/go-redis/v9
    RUN go mod tidy
    
    # Build the binary
    RUN go build -o transcode-worker .
    
    # ---------- Final Image ----------
    FROM debian:bullseye
    
    # Install FFmpeg and CA certificates
    RUN apt-get update && \
        apt-get install -y ffmpeg ca-certificates && \
        rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    