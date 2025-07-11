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
    
    # ✅ Step 1: Clone and build vvenc from specific release
    RUN git clone --branch v1.6.1 https://github.com/fraunhoferhhi/vvenc.git && \
        cd vvenc && mkdir build && cd build && \
        cmake .. -DCMAKE_BUILD_TYPE=Release && \
        make -j$(nproc) && make install
    
    # ✅ Step 2: Manually install libvvenc.pc and validate existence
    RUN mkdir -p /usr/local/lib/pkgconfig && \
        echo -e "prefix=/usr/local\n\
    exec_prefix=\${prefix}\n\
    libdir=\${exec_prefix}/lib\n\
    includedir=\${prefix}/include\n\
    \n\
    Name: vvenc\n\
    Description: Fraunhofer Versatile Video Encoder\n\
    Version: 1.6.1\n\
    Libs: -L\${libdir} -lvvenc\n\
    Cflags: -I\${includedir}" > /usr/local/lib/pkgconfig/libvvenc.pc && \
        echo "[DEBUG] Confirm libvvenc.pc is created:" && \
        ls -l /usr/local/lib/pkgconfig/libvvenc.pc
    
    # ✅ Step 3: Test if pkg-config can detect it
    ENV PKG_CONFIG_PATH="/usr/local/lib/pkgconfig"
    ENV LD_LIBRARY_PATH="/usr/local/lib:/usr/local/lib64"
    
    RUN echo "[DEBUG] pkg-config search path:" && echo $PKG_CONFIG_PATH && \
        echo "[DEBUG] List all .pc files known to pkg-config:" && \
        pkg-config --list-all | grep vvenc || true && \
        echo "[DEBUG] pkg-config modversion:" && \
        pkg-config --modversion libvvenc && \
        echo "[DEBUG] pkg-config --libs:" && \
        pkg-config --libs libvvenc && \
        echo "[DEBUG] pkg-config --cflags:" && \
        pkg-config --cflags libvvenc
    
    # ✅ Step 4: Build FFmpeg with libvvenc support
    RUN git clone https://github.com/FFmpeg/FFmpeg.git && \
        cd FFmpeg && \
        echo "[DEBUG] Begin FFmpeg configure..." && \
        ./configure \
          --pkg-config-flags="--static" \
          --extra-cflags="-I/usr/local/include" \
          --extra-ldflags="-L/usr/local/lib -L/usr/local/lib64" \
          --enable-gpl --enable-nonfree \
          --enable-libx264 --enable-libx265 --enable-libvvenc && \
        echo "[DEBUG] Configure successful. Start building FFmpeg..." && \
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
    