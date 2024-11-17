# Use the official Golang image as the base image
FROM golang:1.21.5-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o peripheral_server peripheral_server_main/main.go

# Expose the port that the peripheral server listens on
EXPOSE 2001

# Set the entry point for the container
CMD ["./peripheral_server"]
