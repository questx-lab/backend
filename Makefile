ENVIRONMENT ?= dev

DEPLOYMENT_DIR = ${PWD}/deploy/${ENVIRONMENT}

COMPOSE_FILE := ${DEPLOYMENT_DIR}/docker-compose.yml
START_PROXY_FILE := $(DEPLOYMENT_DIR)/start_game_proxy.sh
START_API_FILE := $(DEPLOYMENT_DIR)/start_api.sh
START_PROCESSOR_FILE := $(DEPLOYMENT_DIR)/start_game_processor.sh


build:
	go build -o app-exe ./cmd/srv/.

cert-gen:
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server.key -out server.crt -subj \
		"/C=US/ST=NewYork/L=XX/O=quest-x/OU=questx/CN=questx.com"
	@echo "The output is server.key and server.crt"

gen-mock:
	mockery --all --case underscore

start-db:
	docker compose -f ${COMPOSE_FILE} up mysql -d

start-storage:
	docker compose -f ${COMPOSE_FILE} up minio -d

start-kafka:
	docker compose -f ${COMPOSE_FILE} up kafka -d

start-game-proxy:
	${START_PROXY_FILE}

start-game-processor:
	${START_PROCESSOR_FILE}

start-api:
	${START_API_FILE}

start-redis:
	docker compose -f ${COMPOSE_FILE} up redis -d
