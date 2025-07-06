# ---------- Builder ----------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    RUN go build -o transcode-worker .
    
    # ---------- Final Image ----------
    FROM debian:bullseye
    
    # Install FFmpeg and certificates (with retry/fix)
    RUN apt-get update --fix-missing && \
        apt-get install -y ffmpeg ca-certificates && \
        rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    