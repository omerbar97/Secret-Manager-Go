# Makefile for building and running the Docker image

# Build the Docker image
build:
	docker build -t secret-manager .

# Run the Docker container
run:
	docker run -p 8080:8080 secret-manager

# Remove the Docker container
clean:
	docker stop $$(docker ps -a -q) && docker rm $$(docker ps -a -q)
