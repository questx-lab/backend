#!/bin/bash

while read -ra e; do
  export $e
done <<<"$(cat ./developments/.env)"
go run ./cmd/srv/.
