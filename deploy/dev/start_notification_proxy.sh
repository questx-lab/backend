#!/bin/bash
source ./deploy/dev/export_env.sh
go run ./cmd/srv/. notification_proxy
