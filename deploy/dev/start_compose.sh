#!/bin/bash

while read -ra e; do
  export $e
done <<<"$(cat ./deploy/dev/.env)"
docker compose -f deploy/dev/docker-compose-all.yml up -d
