#!/bin/bash

# Name of the Docker image
IMAGE_NAME="geom_to_geojson"
# Name of the Docker container
CONTAINER_NAME="geom-to-geojson-container"
# Path to your JSON configuration file
CONFIG_JSON_PATH="test-file.json"
# Check if the Docker image already exists
IMAGE_EXISTS=$(docker images -q $IMAGE_NAME)

# Build the Docker image if it doesn't exist
if [ -z "$IMAGE_EXISTS" ]; then
    echo "Building Docker image..."
    docker build -f Dockerfile.geom_to_geojson -t $IMAGE_NAME .
else
    echo "Docker image already exists. Skipping build."
fi
docker build -f Dockerfile.geom_to_geojson -t $IMAGE_NAME .
# Extract JSON from file
JSON_STRING=$(jq -c . "$CONFIG_JSON_PATH")
if [ $? -ne 0 ]; then
    echo "Failed to parse JSON from $CONFIG_JSON_PATH"
    exit 1
fi

# Run the Docker container with the JSON string
docker run -d --name $CONTAINER_NAME $IMAGE_NAME "$JSON_STRING"

# Wait for the container to finish its execution
docker wait $CONTAINER_NAME

# Display the logs of the container
echo "Container logs:"
docker logs $CONTAINER_NAME

# Remove the container after it's done
docker rm $CONTAINER_NAME

echo "Process completed and container removed."



