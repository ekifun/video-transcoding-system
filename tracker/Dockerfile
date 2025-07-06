# tracker/Dockerfile

FROM golang:1.22.3 as builder
WORKDIR /app
COPY . .
RUN go build -o tracker .

FROM debian:bullseye
WORKDIR /app
COPY --from=builder /app/tracker /app/tracker
ENTRYPOINT ["./tracker"]
