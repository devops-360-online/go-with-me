SHELL := /bin/bash

.PHONY: start
start: 
	docker-compose up -d

.PHONY: build
build: 
	docker-compose up -d --build 

.PHONY: init
init:
	mkdir -p logs/app logs/kafka logs/mongodb
	aws --endpoint-url=http://localhost:4566 s3 mb s3://go-with-me-images-events
	aws --endpoint-url=http://localhost:4566 s3 mb s3://go-with-me-images-chat-event
	docker-compose up -d 

.PHONY: destroy-all
destroy-all:
	@echo "You are about to delete all volumes and restart the services."
	@read -p "Are you sure? [yes/no] " CONFIRM; \
	if [[ "$$CONFIRM" == "yes" || "$$CONFIRM" == "y" ]]; then \
		docker-compose down --volumes; \
		yes | docker volume prune; \
		rm -rf logs; \
		echo "All volumes and logs have been deleted."; \
		exit 0; \
	else \
		echo "Operation canceled."; \
		exit 125; \
	fi

.PHONY: restart
restart:
	@bash -c ' \
		if $(MAKE) destroy-all; then \
		   $(MAKE) init; \
		   $(MAKE) build; \
		else \
			echo "Restart operation skipped."; \
		fi \
	'

.PHONY: pgadmin
pgadmin:
	docker run --name pgadmin-container -p 5050:80 -e PGADMIN_DEFAULT_EMAIL=admin@x.x -e PGADMIN_DEFAULT_PASSWORD=admin -d dpage/pgadmin4:8.13.0
