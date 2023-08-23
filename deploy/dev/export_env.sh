#!/bin/bash

while read -ra e; do
  export $e
done <<<"$(cat ./deploy/dev/.env)"

export GOOGLE_AUTHENTICATION_CREDENTIALS_JSON=$(cat google-service.json)
