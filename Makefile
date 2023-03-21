COMPOSE_FILE := ${PWD}/developments/docker-compose.yml
gen-proto:
	docker compose -f ${COMPOSE_FILE} up generate_pb_go 
build:
	go build -o app-exe ./cmd/srv/.

cert-gen:
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server.key -out server.crt -subj \
		"/C=US/ST=NewYork/L=XX/O=quest-x/OU=questx/CN=questx.com"
	@echo "The output is server.key and server.crt"
