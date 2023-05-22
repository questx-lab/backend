#!/bin/bash

while read -ra e; do
  export $e
done <<<"$(cat ./deploy/dev/.env)"
<<<<<<<< HEAD:deploy/dev/start_game_proxy.sh
go run ./cmd/srv/. game_proxy
========
go run ./cmd/srv/. cron
>>>>>>>> 38b8adb36ebc96c039d3f7260a332b9e9326889a:deploy/dev/start_cron.sh
