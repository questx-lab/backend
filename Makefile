ENVIRONMENT ?= dev

DEPLOYMENT_DIR = ${PWD}/deploy/${ENVIRONMENT}

COMPOSE_FILE := ${DEPLOYMENT_DIR}/docker-compose.yml
START_PROXY_FILE := $(DEPLOYMENT_DIR)/start_game_proxy.sh
START_API_FILE := $(DEPLOYMENT_DIR)/start_api.sh
START_ENGINE_FILE := $(DEPLOYMENT_DIR)/start_game_engine.sh
START_COMPOSE_FILE := $(DEPLOYMENT_DIR)/start_compose.sh

build:
	go build -o app ./cmd/srv/.

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

start-game-engine:
	${START_ENGINE_FILE}

start-api:
	${START_API_FILE}

docker-build:
	docker build -t questx -f deploy/Dockerfile .

start-compose:
	${START_COMPOSE_FILE}

stop-compose:
	docker compose -f deploy/dev/docker-compose-all.yml down

start-redis:
	docker compose -f ${COMPOSE_FILE} up redis -d
