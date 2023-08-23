#!/bin/bash
source ./deploy/dev/export_env.sh
docker compose -f deploy/dev/docker-compose-all.yml up -d
