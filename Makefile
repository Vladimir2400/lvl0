.PHONY: test test-cover test-verbose run build clean docker-up docker-down lint

# Запуск всех тестов
test:
	go test ./...

# Запуск тестов с покрытием кода
test-cover:
	go test -v -cover ./...

# Запуск тестов с детальным выводом и покрытием
test-verbose:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Запуск тестов конкретного пакета
test-models:
	go test -v ./models/

test-config:
	go test -v ./config/

test-cache:
	go test -v ./internal/cache/

test-validator:
	go test -v ./internal/validator/

test-repository:
	go test -v ./internal/repository/

test-generator:
	go test -v ./cmd/generator/

test-main:
	go test -v .

# Запуск основного приложения
run:
	go run main.go

# Генерация тестовых данных
generate-data:
	go run cmd/generator/main.go -count=10 -delay=1s

# Сборка приложения
build:
	go build -o bin/wb-service main.go
	go build -o bin/generator cmd/generator/main.go

# Очистка
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Запуск Docker Compose (PostgreSQL, Kafka)
docker-up:
	docker-compose up -d

# Остановка Docker Compose
docker-down:
	docker-compose down

# Проверка статуса контейнеров
docker-status:
	docker-compose ps

# Линтер (если установлен golangci-lint)
lint:
	golangci-lint run

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Benchmark тесты
benchmark:
	go test -bench=. -benchmem ./internal/cache/

# Профилирование CPU
profile-cpu:
	go test -cpuprofile=cpu.prof -bench=. ./internal/cache/

# Профилирование памяти
profile-mem:
	go test -memprofile=mem.prof -bench=. ./internal/cache/

# Проверка покрытия тестами (минимум 80%)
test-coverage-check:
	@echo "Checking test coverage..."
	@go test -coverprofile=coverage.out ./... > /dev/null
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < 80" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage $$COVERAGE% is below 80%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$COVERAGE% meets the 80% requirement"; \
	fi

# Интеграционные тесты (требуют запущенную инфраструктуру)
test-integration:
	@echo "Starting integration tests..."
	@docker-compose up -d
	@sleep 10
	@go test -tags=integration ./...
	@docker-compose down

# Помощь
help:
	@echo "Available commands:"
	@echo "  test                  - Run all tests"
	@echo "  test-cover           - Run tests with coverage"
	@echo "  test-verbose         - Run tests with detailed output and HTML coverage report"
	@echo "  test-coverage-check  - Check if coverage meets 80% requirement"
	@echo "  test-{package}       - Run tests for specific package"
	@echo ""
	@echo "  run                  - Run the main application"
	@echo "  generate-data        - Generate test data for Kafka"
	@echo "  build                - Build applications"
	@echo "  clean                - Clean build artifacts"
	@echo ""
	@echo "  docker-up            - Start Docker Compose services"
	@echo "  docker-down          - Stop Docker Compose services"
	@echo "  docker-status        - Check Docker Compose status"
	@echo ""
	@echo "  lint                 - Run linter"
	@echo "  deps                 - Install dependencies"
	@echo "  benchmark            - Run benchmark tests"
	@echo "  test-integration     - Run integration tests"