# ---------- Stage 1: Build Go App ----------
    FROM golang:1.22.3-bullseye AS builder

    WORKDIR /app
    COPY . .
    
    RUN go get github.com/redis/go-redis/v9
    RUN go mod tidy
    RUN go build -o transcode-worker .
    
    # ---------- Stage 2: Build FFmpeg Dependencies ----------
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
    
    # Step 1: Clone and build libvvenc
    RUN git clone --branch v1.6.1 https://github.com/fraunhoferhhi/vvenc.git && \
        cd vvenc && mkdir build && cd build && \
        cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/opt/vvenc/install && \
        make -j$(nproc) && make install
    
    # Step 2: Install vvenc headers and libs
    RUN mkdir -p /usr/local/include/vvenc && \
        cp -r /opt/vvenc/install/include/vvenc/* /usr/local/include/vvenc/ && \
        cp -r /opt/vvenc/install/lib/libvvenc* /usr/local/lib/ && \
        ldconfig
    
    # Step 3: Create pkg-config file using safe heredoc block
    RUN mkdir -p /usr/local/lib/pkgconfig && \
        sh -c 'cat > /usr/local/lib/pkgconfig/libvvenc.pc <<EOF
    prefix=/usr/local
    exec_prefix=\${prefix}
    libdir=\${exec_prefix}/lib
    includedir=\${prefix}/include
    
    Name: vvenc
    Description: Fraunhofer Versatile Video Encoder
    Version: 1.6.1
    Libs: -L\${libdir} -lvvenc -lstdc++
    Cflags: -I\${includedir}
    EOF'
    
    # Set environment variables for FFmpeg build
    ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
    ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
    ENV LDFLAGS="-lstdc++"
    ENV CXX=g++
    ENV CC=gcc
    
    # ---------- Stage 3: Debug-Only Container for Interactive FFmpeg Build ----------
    FROM ffmpeg-builder AS debug-ffmpeg
    WORKDIR /opt
    CMD ["/bin/bash"]
    
    # ---------- Stage 4: Build FFmpeg ----------
    FROM ffmpeg-builder AS ffmpeg-builder-final
    
    WORKDIR /opt
    RUN git clone https://github.com/FFmpeg/FFmpeg.git && \
        cd FFmpeg && \
        ./configure \
          --pkg-config-flags="--static" \
          --extra-cflags="$(pkg-config --cflags libvvenc)" \
          --extra-ldflags="$(pkg-config --libs libvvenc)" \
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
    
    # Copy FFmpeg and vvenc libs
    COPY --from=ffmpeg-builder-final /usr/local/bin/ffmpeg /usr/local/bin/
    COPY --from=ffmpeg-builder-final /usr/local/bin/ffprobe /usr/local/bin/
    COPY --from=ffmpeg-builder-final /usr/local/lib/libvvenc* /usr/local/lib/
    COPY --from=ffmpeg-builder-final /usr/local/lib64/libvvenc* /usr/local/lib64/ || true
    
    # Copy the Go binary
    WORKDIR /app
    COPY --from=builder /app/transcode-worker /app/transcode-worker
    
    ENTRYPOINT ["./transcode-worker"]
    