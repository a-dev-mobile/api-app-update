# Имя файла docker-compose
DOCKER_COMPOSE_FILE=docker/docker-compose.yml

# Сборка и запуск контейнеров
.PHONY: up
up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d

# Остановка и удаление контейнеров, а также удаление тома данных
.PHONY: down
down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@if docker volume inspect docker_postgres_data > /dev/null 2>&1; then \
		echo "Removing volume docker_postgres_data"; \
		docker volume rm docker_postgres_data; \
	else \
		echo "Volume docker_postgres_data does not exist"; \
	fi

# Пересборка контейнеров
.PHONY: rebuild
rebuild: down
	$(MAKE) up

# Вывод логов
.PHONY: logs
logs:
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# Открытие оболочки в контейнере postgres
.PHONY: shell
shell:
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec postgres sh

# Очистка неиспользуемых данных Docker
.PHONY: clean
clean:
	docker system prune -f
	docker volume prune -f

# Просмотр состояния контейнеров
.PHONY: ps
ps:
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps

# Проверка и удаление неиспользуемых образов, контейнеров и т.д.
.PHONY: prune
prune:
	docker system prune -f
	docker volume prune -f
	docker network prune -f

# Перезапуск контейнеров
.PHONY: restart
restart:
	docker-compose -f $(DOCKER_COMPOSE_FILE) restart
