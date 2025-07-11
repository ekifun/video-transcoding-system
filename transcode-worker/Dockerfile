# ---------- Stage 1: Build Go App ----------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    # Install Go dependencies and build
    RUN go get github.com/redis/go-redis/v9
    RUN go mod tidy
    RUN go build -o transcode-worker .
    
    
    # ---------- Stage 2: Build FFmpeg with libvvenc ----------
    FROM debian:bullseye AS ffmpeg-builder
    
    ENV DEBIAN_FRONTEND=noninteractive
    
    # Install build dependencies
    RUN apt-get update && apt-get install -y \
      git cmake build-essential yasm pkg-config \
      wget curl unzip nasm libx264-dev libx265-dev libnuma-dev \
      zlib1g-dev libass-dev libfreetype6-dev libvorbis-dev \
      libxcb1-dev libxcb-shm0-dev libxcb-xfixes0-dev libtool automake \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*
    
    WORKDIR /opt
    
    # Build and install libvvenc
    RUN git clone https://github.com/fraunhoferhhi/vvenc.git && \
        cd vvenc && mkdir build && cd build && \
        cmake .. -DCMAKE_BUILD_TYPE=Release && \
        make -j$(nproc) && make install
    
    # Fix pkg-config path so FFmpeg detects libvvenc >= 1.6.1
    ENV PKG_CONFIG_PATH="/usr/local/lib64/pkgconfig:/usr/local/lib/pkgconfig"
    
    # Build and install FFmpeg with VVC support
    RUN git clone https://github.com/FFmpeg/FFmpeg.git && \
        cd FFmpeg && \
        ./configure --enable-gpl --enable-nonfree --enable-libx264 --enable-libx265 --enable-libvvenc && \
        make -j$(nproc) && make install
    
    
    # ---------- Final Stage: Runtime Image ----------
    FROM debian:bullseye
    
    ENV DEBIAN_FRONTEND=noninteractive
    ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
    
    # Install runtime dependencies
    RUN apt-get update && apt-get install -y \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*
    
    # Copy FFmpeg binaries and VVC libraries
    COPY --from=ffmpeg-builder /usr/local/bin/ffmpeg /usr/local/bin/
    COPY --from=ffmpeg-builder /usr/local/bin/ffprobe /usr/local/bin/
    COPY --from=ffmpeg-builder /usr/local/lib/libvvenc* /usr/local/lib/
    COPY --from=ffmpeg-builder /usr/local/lib64/libvvenc* /usr/local/lib64/
    
    # Copy transcode-worker Go binary
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    