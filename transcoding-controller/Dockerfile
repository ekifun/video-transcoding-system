# Stage 1: Build the Go binary
FROM golang:1.22.3-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY . .

RUN go build -o transcoding-controller

# Stage 2: Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/transcoding-controller .

EXPOSE 8080
ENTRYPOINT ["./transcoding-controller"]
