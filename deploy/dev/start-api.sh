#!/bin/bash

while read -ra e; do
  export $e
done <<<"$(cat ./deploy/dev/.env)"
go run ./cmd/srv/. api
