COMPOSE_FILE := ${PWD}/developments/docker-compose.yml
gen-proto:
	docker compose -f ${COMPOSE_FILE} up generate_pb_go 
build:
	go build -o app-exe ./cmd/srv/.
gen-mock:
	mockery --all --case underscore
start-mysql:
	docker compose -f ${COMPOSE_FILE} up mysql -d