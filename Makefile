ENVIRONMENT ?= dev

DEPLOYMENT_DIR = ${PWD}/deploy/${ENVIRONMENT}

COMPOSE_FILE := ${DEPLOYMENT_DIR}/docker-compose.yml
START_API_FILE := $(DEPLOYMENT_DIR)/start_api.sh
START_CRON_FILE := $(DEPLOYMENT_DIR)/start_cron.sh
START_SEARCH_FILE := $(DEPLOYMENT_DIR)/start_search.sh
START_BLOCKCHAIN_FILE := $(DEPLOYMENT_DIR)/start_blockchain.sh
START_NOTIFICATION_PROXY := $(DEPLOYMENT_DIR)/start_notification_proxy.sh
START_NOTIFICATION_ENGINE := $(DEPLOYMENT_DIR)/start_notification_engine.sh
START_COMPOSE_FILE := $(DEPLOYMENT_DIR)/start_compose.sh

contract-gen:
	solc --abi contract/erc20.sol --overwrite -o contract/erc20
	solc --bin contract/erc20.sol --overwrite -o contract/erc20
	abigen --bin=contract/erc20/IERC20Metadata.bin --abi=contract/erc20/IERC20Metadata.abi --pkg=contract --out=contract/erc20.go

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
	docker compose -f ${COMPOSE_FILE} up adminer -d

start-storage:
	docker compose -f ${COMPOSE_FILE} up minio -d

start-kafka:
	docker compose -f ${COMPOSE_FILE} up kafka -d

start-blockchain:
	${START_BLOCKCHAIN_FILE}

start-api:
	${START_API_FILE}

start-cron:
	${START_CRON_FILE}

start-search:
	${START_SEARCH_FILE}

start-notification-proxy:
	${START_NOTIFICATION_PROXY}

start-notification-engine:
	${START_NOTIFICATION_ENGINE}

docker-build:
	docker build -t questx -f deploy/Dockerfile .

start-compose:
	${START_COMPOSE_FILE}

stop-compose:
	docker compose -f deploy/dev/docker-compose-all.yml down

start-redis:
	docker compose -f ${COMPOSE_FILE} up redis -d

start-scylladb:
	docker compose -f ${COMPOSE_FILE} up scylladb -d
