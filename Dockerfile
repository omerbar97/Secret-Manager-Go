# Use the official golang image as the base image
FROM golang:1.21.1

# Set the working directory inside the container
WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY api/ ./api/
COPY types/ ./types/
COPY utils/ ./utils/

RUN go mod download

# Build the CLI and server binaries
RUN go build -o ./bin/server ./api/server

# Expose the necessary port for the server
EXPOSE 8080

# Command to run the server when the container starts
CMD ["./bin/server"]
