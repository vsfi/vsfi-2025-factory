.PHONY: test test-verbose test-coverage test-models test-config test-services test-clean build run help

# Цвета для вывода
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

# Запуск всех тестов
test:
	@echo "$(GREEN)🧪 Запуск всех unit-тестов...$(RESET)"
	go test ./...

# Запуск тестов с подробным выводом
test-verbose:
	@echo "$(GREEN)🧪 Запуск тестов с подробным выводом...$(RESET)"
	go test -v ./...

# Запуск тестов с покрытием кода
test-coverage:
	@echo "$(GREEN)📊 Анализ покрытия кода тестами...$(RESET)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(BLUE)📈 Отчет о покрытии сохранен в coverage.html$(RESET)"

# Тестирование конкретных модулей
test-models:
	@echo "$(GREEN)🏗️  Тестирование моделей данных...$(RESET)"
	go test -v ./internal/models

test-config:
	@echo "$(GREEN)⚙️  Тестирование конфигурации...$(RESET)"
	go test -v ./internal/config

test-services:
	@echo "$(GREEN)🔧 Тестирование сервисов...$(RESET)"
	go test -v ./internal/services

# Очистка тестовых файлов
test-clean:
	@echo "$(YELLOW)🧹 Очистка тестовых файлов...$(RESET)"
	rm -f coverage.out coverage.html

# Сборка проекта
build:
	@echo "$(GREEN)🔨 Сборка factory...$(RESET)"
	go build -o factory cmd/main.go

# Запуск в режиме разработки
dev:
	@echo "$(GREEN)🚀 Запуск в режиме разработки...$(RESET)"
	LOG_LEVEL=debug go run cmd/main.go

# Запуск продакшн версии
run: build
	@echo "$(GREEN)🚀 Запуск factory...$(RESET)"
	./factory

# Установка зависимостей
deps:
	@echo "$(GREEN)📦 Установка зависимостей...$(RESET)"
	go mod download
	go mod tidy

# Линтинг кода
lint:
	@echo "$(GREEN)🔍 Проверка кода линтером...$(RESET)"
	golangci-lint run

# Форматирование кода
fmt:
	@echo "$(GREEN)✨ Форматирование кода...$(RESET)"
	go fmt ./...

# Полная проверка (тесты + линтинг + форматирование)
check: fmt lint test

# Справка
help:
	@echo "$(BLUE)🛠️  Factory Service - Доступные команды:$(RESET)"
	@echo ""
	@echo "$(GREEN)Тестирование:$(RESET)"
	@echo "  make test          - Запуск всех unit-тестов"
	@echo "  make test-verbose  - Запуск тестов с подробным выводом"
	@echo "  make test-coverage - Анализ покрытия кода тестами"
	@echo "  make test-models   - Тестирование только моделей"
	@echo "  make test-config   - Тестирование только конфигурации"
	@echo "  make test-services - Тестирование только сервисов"
	@echo "  make test-clean    - Очистка файлов тестирования"
	@echo ""
	@echo "$(GREEN)Разработка:$(RESET)"
	@echo "  make build         - Сборка проекта"
	@echo "  make dev           - Запуск в режиме разработки"
	@echo "  make run           - Запуск продакшн версии"
	@echo "  make deps          - Установка зависимостей"
	@echo ""
	@echo "$(GREEN)Качество кода:$(RESET)"
	@echo "  make fmt           - Форматирование кода"
	@echo "  make lint          - Проверка линтером"
	@echo "  make check         - Полная проверка (fmt + lint + test)"
	@echo ""
	@echo "$(GREEN)Справка:$(RESET)"
	@echo "  make help          - Эта справка" 