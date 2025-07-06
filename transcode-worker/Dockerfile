# ---------- Builder ----------
    FROM golang:1.21-bullseye AS builder

    WORKDIR /app
    COPY . .
    RUN go build -o transcode-worker .
    
    # ---------- Final Image ----------
    FROM debian:bullseye
    
    RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    