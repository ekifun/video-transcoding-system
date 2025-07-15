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
  
  RUN apt-get update && apt-get install -y \
    git cmake build-essential yasm pkg-config \
    wget curl unzip nasm libnuma-dev \
    zlib1g-dev libass-dev libfreetype6-dev libvorbis-dev \
    libxcb1-dev libxcb-shm0-dev libxcb-xfixes0-dev libxcb-shape0-dev \
    libtool automake ca-certificates && \
    rm -rf /var/lib/apt/lists/*
  
  WORKDIR /opt
  
  # Build libx264 (shared)
  RUN git clone https://code.videolan.org/videolan/x264.git && \
    cd x264 && \
    ./configure --enable-shared --prefix=/usr/local && \
    make -j$(nproc) && make install
  
  # Build libx265 (shared)
  RUN git clone https://bitbucket.org/multicoreware/x265_git && \
    cd x265_git/build/linux && \
    cmake -G "Unix Makefiles" -DENABLE_SHARED=ON ../../source && \
    make -j$(nproc) && make install
  
  # Build libvpx (with multithread and row-mt support)
  RUN git clone https://chromium.googlesource.com/webm/libvpx && \
    cd libvpx && \
    ./configure --prefix=/usr/local --enable-shared --disable-static \
      --enable-multithread --enable-row-mt --disable-examples --disable-unit-tests && \
    make -j$(nproc) && make install
  
  # Build libvvenc (shared)
  RUN git clone --branch v1.6.1 https://github.com/fraunhoferhhi/vvenc.git && \
      cd vvenc && mkdir build && cd build && \
      cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/opt/vvenc/install && \
      make -j$(nproc) && make install
  
  RUN mkdir -p /usr/local/include/vvenc && \
      cp -r /opt/vvenc/install/include/vvenc/* /usr/local/include/vvenc/ && \
      cp -r /opt/vvenc/install/lib/libvvenc* /usr/local/lib/ && \
      mkdir -p /usr/local/lib64 && \
      cp -r /opt/vvenc/install/lib64/libvvenc* /usr/local/lib64/ 2>/dev/null || true && \
      ldconfig
  
  RUN mkdir -p /usr/local/lib/pkgconfig && \
      echo "prefix=/usr/local\n\
  exec_prefix=\${prefix}\n\
  libdir=\${exec_prefix}/lib\n\
  includedir=\${prefix}/include\n\
  \n\
  Name: vvenc\n\
  Description: Fraunhofer Versatile Video Encoder\n\
  Version: 1.6.1\n\
  Libs: -L\${libdir} -lvvenc -lstdc++\n\
  Cflags: -I\${includedir}" > /usr/local/lib/pkgconfig/libvvenc.pc
  
  ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
  ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
  ENV LDFLAGS="-lstdc++"
  ENV CXX=g++
  ENV CC=gcc
  
  # ---------- Stage 3: Build FFmpeg ----------
  FROM ffmpeg-builder AS ffmpeg-builder-final
  
  WORKDIR /opt
  RUN git clone https://github.com/FFmpeg/FFmpeg.git && \
      cd FFmpeg && \
      ./configure \
        --pkg-config-flags="--static" \
        --extra-cflags="$(pkg-config --cflags libvvenc)" \
        --extra-ldflags="$(pkg-config --libs libvvenc)" \
        --enable-gpl --enable-nonfree \
        --enable-libx264 --enable-libx265 --enable-libvvenc \
        --enable-libvpx \
        --enable-shared && \
      make -j$(nproc) && make install
  
  # ---------- Stage 4: Runtime Image ----------
  FROM debian:bullseye
  
  ENV DEBIAN_FRONTEND=noninteractive
  ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64
  
  RUN apt-get update && apt-get install -y \
    ca-certificates \
    libxcb1 \
    libxcb-shm0 \
    libxcb-xfixes0 \
    libxcb-shape0 \
    libnuma1 \
    libass9 \
    libfreetype6 \
    libvorbis0a \
    libvorbisenc2 \
    zlib1g \
    && rm -rf /var/lib/apt/lists/*
  
  COPY --from=ffmpeg-builder-final /usr/local/bin/ffmpeg /usr/local/bin/
  COPY --from=ffmpeg-builder-final /usr/local/bin/ffprobe /usr/local/bin/
  COPY --from=ffmpeg-builder-final /usr/local/lib/ /usr/local/lib/
  COPY --from=ffmpeg-builder-final /usr/local/lib64/ /usr/local/lib64/
  COPY --from=ffmpeg-builder-final /usr/local/lib/libx264.so* /usr/local/lib/
  COPY --from=ffmpeg-builder-final /usr/local/lib/libx265.so* /usr/local/lib/
  COPY --from=ffmpeg-builder-final /usr/local/lib/libvpx.so* /usr/local/lib/
  
  RUN ldconfig
  
  WORKDIR /app
  COPY --from=builder /app/transcode-worker /app/transcode-worker
  
  ENTRYPOINT ["./transcode-worker"]
  