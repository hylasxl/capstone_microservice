# Build Stage
FROM golang:1.23.3-alpine3.20 AS build

# Set the Current Working Directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy source code and credentials
COPY . .
COPY syncio-7a920-firebase-adminsdk-3esie-1b79ddb11e.json /app/syncio-7a920-firebase-adminsdk-3esie-1b79ddb11e.json

# Build the Go app
RUN go build -o notification_service

# Final Stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory
WORKDIR /root/

# Copy the compiled binary and Firebase credentials
COPY --from=build /app/notification_service .
COPY --from=build /app/syncio-7a920-firebase-adminsdk-3esie-1b79ddb11e.json ./syncio-7a920-firebase-adminsdk-3esie-1b79ddb11e.json

# Expose the gRPC port
EXPOSE 50057

# Set environment variable for Firebase credentials
ENV FIREBASE_CREDENTIALS=/root/syncio-7a920-firebase-adminsdk-3esie-1b79ddb11e.json

# Command to run the application
CMD ["./notification_service"]
