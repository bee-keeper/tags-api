# Use the official Golang image
FROM golang:1.23.1

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Install air
RUN go install github.com/air-verse/air@latest

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Command to run when the container starts
CMD ["air"]
