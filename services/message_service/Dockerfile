# Use the official Golang image as the base image
FROM golang:1.23.3-alpine3.20 AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o message_service

# Start a new stage from a smaller image to reduce the image size
FROM alpine:latest

# Install necessary packages for running the Go app
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the binary from the build stage
COPY --from=build /app/message_service .

# Expose port 50051 for the gRPC service
EXPOSE 50055

# Command to run the executable
CMD ["./message_service"]
