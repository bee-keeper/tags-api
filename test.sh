#!/bin/bash

CONTAINER_ID=$(docker ps -q -f "name=tags-api")

if [ -z "$CONTAINER_ID" ]; then
  echo "Container 'tags-api' not found."
  exit 1
fi

docker exec "$CONTAINER_ID" /bin/bash -c "POSTGRES_DB=test && go test ./... -cover -v"
