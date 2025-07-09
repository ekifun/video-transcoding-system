# -------- Build Stage --------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    # Fetch dependencies
    RUN go mod tidy
    
    # Build the Go binary
    RUN go build -o mpd-generator .
    
    # -------- Runtime Stage --------
    FROM debian:bullseye
    
    # Install MP4Box (GPAC) and SQLite dependency
    RUN apt-get update && \
        apt-get install -y --no-install-recommends \
            gpac \
            libsqlite3-0 && \
        rm -rf /var/lib/apt/lists/*
    
    # Setup working directory
    WORKDIR /app
    
    # Copy binary from builder stage
    COPY --from=builder /app/mpd-generator /app/mpd-generator
    
    # Copy any static DB initialization file if needed
    # COPY --from=builder /app/db/data/jobs.db /app/db/data/jobs.db
    
    # Expose volume if SQLite DB is persisted
    VOLUME ["/app/db/data"]
    
    # Entrypoint
    ENTRYPOINT ["./mpd-generator"]
    