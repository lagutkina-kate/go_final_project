APP_NAME := go-final-project
IMAGE_NAME := go-final-project
CONTAINER_NAME := go-final-project
PORT := 7540
DOCKER := sudo docker

.PHONY: test docker-run

test:
	@set -e; \
	go build -o $(APP_NAME) .; \
	TODO_PORT=:$(PORT) TODO_DBFILE=scheduler.db TODO_PASSWORD=12345 ./$(APP_NAME) & \
	APP_PID=$$!; \
	trap 'kill $$APP_PID >/dev/null 2>&1 || true' EXIT; \
	echo "Ожидание запуска сервера на порту $(PORT)..."; \
	
	TODO_PORT=$(PORT) go test -v ./tests/...

docker-run:
	$(DOCKER) build -t $(IMAGE_NAME) .
	-$(DOCKER) stop $(CONTAINER_NAME)
	-$(DOCKER) rm $(CONTAINER_NAME)
	$(DOCKER) run -d \
		--name $(CONTAINER_NAME) \
		-p $(PORT):7540 \
		-v $(CURDIR)/scheduler.db:/app/scheduler.db \
		$(IMAGE_NAME)
