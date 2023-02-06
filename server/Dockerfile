# Use an official Golang image as the base image
FROM golang:latest

# Set the working directory to /app
WORKDIR /app

# Copy the entire project to the container
COPY . .

# Install the dependencies
RUN go get -d -v ./...

# Build the go-chi application
RUN go build -o main .

# Expose port 13337 to the host
EXPOSE 13337

# Set the command to run the go-chi application
CMD ["./main"]