# Use the official Go image as a base
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project into the working directory
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o gitback main.go

# Start a new stage from scratch
FROM alpine:latest

# Install git
RUN apk add --no-cache git

# Set the working directory for the final image
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/gitback .

# Command to run the executable
ENTRYPOINT ["/root/gitback"]
