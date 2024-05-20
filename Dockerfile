# Use the official Golang image as the base image
FROM golang:1.22-bookworm

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o gh-proxy-go

# Expose the port that the application listens on
EXPOSE 8001

# Set the command to run the application
CMD ["./gh-proxy-go"]