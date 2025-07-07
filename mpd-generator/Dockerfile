# -------- Build Stage --------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    # Optional: tidy and fetch dependencies
    RUN go mod tidy
    RUN go build -o mpd-generator .
    
    # -------- Runtime Stage --------
    FROM debian:bullseye
    
    # Install MP4Box (GPAC)
    RUN apt-get update && \
    apt-get install -y --no-install-recommends gpac || \
    (apt-get update && apt-get install -y --no-install-recommends gpac) && \
    rm -rf /var/lib/apt/lists/*


    
    WORKDIR /app
    
    # Copy binary from builder stage
    COPY --from=builder /app/mpd-generator /app/mpd-generator
    
    ENTRYPOINT ["./mpd-generator"]
    