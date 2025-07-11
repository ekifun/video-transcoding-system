# ---------- Stage 1: Build Go App ----------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    RUN go get github.com/redis/go-redis/v9
    RUN go mod tidy
    RUN go build -o transcode-worker .
    
    
    # ---------- Stage 2: Build FFmpeg with libvvenc ----------
    FROM debian:bullseye AS ffmpeg-builder
    
    ENV DEBIAN_FRONTEND=noninteractive
    
    RUN apt-get update && apt-get install -y \
      git cmake build-essential yasm pkg-config \
      wget curl unzip nasm libx264-dev libx265-dev libnuma-dev \
      zlib1g-dev libass-dev libfreetype6-dev libvorbis-dev \
      libxcb1-dev libxcb-shm0-dev libxcb-xfixes0-dev libtool automake \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*
    
    WORKDIR /opt
    
    # Build vvenc
    RUN git clone https://github.com/fraunhoferhhi/vvenc.git && \
        cd vvenc && mkdir build && cd build && \
        cmake .. -DCMAKE_BUILD_TYPE=Release && \
        make -j$(nproc) && make install
    
    # Build FFmpeg with libvvenc
    RUN git clone https://github.com/FFmpeg/FFmpeg.git && \
        cd FFmpeg && \
        ./configure --enable-gpl --enable-nonfree --enable-libx264 --enable-libx265 --enable-libvvenc && \
        make -j$(nproc) && make install
    
    
    # ---------- Final Image ----------
    FROM debian:bullseye
    
    ENV LD_LIBRARY_PATH=/usr/local/lib
    ENV DEBIAN_FRONTEND=noninteractive
    
    RUN apt-get update && apt-get install -y \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*
    
    # Copy FFmpeg and libvvenc
    COPY --from=ffmpeg-builder /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
    COPY --from=ffmpeg-builder /usr/local/bin/ffprobe /usr/local/bin/ffprobe
    COPY --from=ffmpeg-builder /usr/local/lib/libvvenc* /usr/local/lib/
    
    # Copy transcode-worker binary
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    