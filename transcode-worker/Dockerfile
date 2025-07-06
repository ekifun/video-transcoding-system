# transcode-worker/Dockerfile
FROM golang:1.22.3-bullseye AS builder

WORKDIR /app
COPY . .

# ✅ Add dependencies
RUN go mod tidy
RUN go get github.com/redis/go-redis/v9

# ✅ Build the binary
RUN go build -o transcode-worker .

FROM debian:bullseye
WORKDIR /app
COPY --from=builder /app/transcode-worker /app/transcode-worker
ENTRYPOINT ["./transcode-worker"]
