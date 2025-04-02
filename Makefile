
.PHONY: start
start:
	@echo "Starting application"
	@docker compose up

.PHONY: start-detached
start-detached:
	@echo "Starting application in detached mode"
	@docker compose up -d

.PHONY: stop
stop:
	@echo "Stopping application"
	@docker compose down

.PHONY: build
build:
	@echo "Building application"
	@go build -o build/app
