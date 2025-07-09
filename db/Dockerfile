# Use a minimal Go base image
FROM golang:1.21-alpine

# Set environment variables
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOFLAGS=-mod=mod

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o db-app ./db/main.go

# Run the DB utility
CMD ["./db-app"]
