# ---------- Builder ----------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    # Optional: Tidy up modules before build (recommended)
    RUN go mod tidy && \
        go build -o transcode-worker .
    
    # ---------- Final Image ----------
    FROM debian:bullseye
    
    # Install required system dependencies (e.g., for ffmpeg if needed)
    # RUN apt-get update && apt-get install -y ffmpeg  # Uncomment if needed
    
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    