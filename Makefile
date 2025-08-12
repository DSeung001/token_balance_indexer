SHELL := /bin/bash
ENV ?= .env
include $(ENV)
export $(shell sed 's/=.*//' $(ENV))

QUEUE_NAME ?= token-events
ACCOUNT_ID ?= 000000000000
SQS_URL = http://localhost:$(LOCALSTACK_EDGE_PORT)
PSQL = psql "host=localhost port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=disable"

.PHONY: up down logs ps psql create-queue list-queue purge-queue schema

up:
	docker compose --env-file $(ENV) up -d

down:
	docker compose --env-file $(ENV) down -v

logs:
	docker compose logs -f --tail=200

ps:
	docker compose ps

psql:
	@$(PSQL)

create-queue:
	aws --endpoint-url=$(SQS_URL) sqs create-queue --queue-name $(QUEUE_NAME)
	@echo "Created queue at: $(SQS_URL)/$(ACCOUNT_ID)/$(QUEUE_NAME)"

list-queue:
	aws --endpoint-url=$(SQS_URL) sqs list-queues

purge-queue:
	aws --endpoint-url=$(SQS_URL) sqs purge-queue --queue-url $(SQS_URL)/$(ACCOUNT_ID)/$(QUEUE_NAME)

schema:
	@cat schema/schema.sql | $(PSQL)
