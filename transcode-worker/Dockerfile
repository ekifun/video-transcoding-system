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
    
    # Install build dependencies
    RUN apt-get update && apt-get install -y \
      git cmake build-essential yasm pkg-config \
      wget curl unzip nasm libx264-dev libx265-dev libnuma-dev \
      zlib1g-dev libass-dev libfreetype6-dev libvorbis-dev \
      libxcb1-dev libxcb-shm0-dev libxcb-xfixes0-dev libtool automake \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*
    
    WORKDIR /opt
    
    # Step 1: Clone and build vvenc (>= v1.6.1)
    RUN git clone --branch v1.6.1 https://github.com/fraunhoferhhi/vvenc.git && \
        cd vvenc && mkdir build && cd build && \
        cmake .. -DCMAKE_BUILD_TYPE=Release && \
        make -j$(nproc) && make install
    
    # Step 2: Manually create pkg-config file for libvvenc
    RUN mkdir -p /usr/local/lib/pkgconfig && \
        echo "prefix=/usr/local" > /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "exec_prefix=\${prefix}" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "libdir=\${exec_prefix}/lib" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "includedir=\${prefix}/include" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "Name: vvenc" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "Description: Fraunhofer Versatile Video Encoder" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "Version: 1.6.1" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "Libs: -L\${libdir} -lvvenc" >> /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "Cflags: -I\${includedir}" >> /usr/local/lib/pkgconfig/libvvenc.pc
    
    ENV PKG_CONFIG_PATH="/usr/local/lib/pkgconfig"
    ENV LD_LIBRARY_PATH="/usr/local/lib:/usr/local/lib64"
    
    # Step 3: Debug info
    RUN echo "[DEBUG] libvvenc.pc contents:" && cat /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "[DEBUG] pkg-config --modversion libvvenc:" && pkg-config --modversion libvvenc && \
        echo "[DEBUG] pkg-config --cflags libvvenc:" && pkg-config --cflags libvvenc && \
        echo "[DEBUG] pkg-config --libs libvvenc:" && pkg-config --libs libvvenc
    
    # Step 4: Clone and build FFmpeg with libvvenc
    RUN git clone https://github.com/FFmpeg/FFmpeg.git && \
        cd FFmpeg && \
        ./configure \
          --pkg-config-flags="--static" \
          --extra-cflags="-I/usr/local/include" \
          --extra-ldflags="-L/usr/local/lib -L/usr/local/lib64" \
          --enable-gpl --enable-nonfree \
          --enable-libx264 --enable-libx265 --enable-libvvenc && \
        make -j$(nproc) && make install
    
    # ---------- Final Stage: Runtime Image ----------
    FROM debian:bullseye
    
    ENV DEBIAN_FRONTEND=noninteractive
    ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
    
    RUN apt-get update && apt-get install -y \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*
    
    COPY --from=ffmpeg-builder /usr/local/bin/ffmpeg /usr/local/bin/
    COPY --from=ffmpeg-builder /usr/local/bin/ffprobe /usr/local/bin/
    COPY --from=ffmpeg-builder /usr/local/lib/libvvenc* /usr/local/lib/
    COPY --from=ffmpeg-builder /usr/local/lib64/libvvenc* /usr/local/lib64/
    
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    