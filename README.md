# Plumbus Factory 🛸

Rick & Morty стилизованный веб-сервис для создания плюмусов с цифровыми подписями, событийной архитектурой и структурированным логированием.

## Описание

Factory - это Go веб-сервис, который предоставляет пользовательский интерфейс для создания плюмбусов с цифровыми подписями. Проект включает полную интеграцию с микросервисной архитектурой:

- 🔐 Авторизация через Keycloak
- ✍️ Цифровые подписи изображений (RSA-SHA256)
- 📨 Событийная архитектура через NATS JetStream
- 📊 Структурированное JSON логирование
- 🎨 Современный UI в стилистике Rick & Morty
- 📈 Прогресс-бар генерации с real-time обновлениями
- 💾 Персональное хранилище изображений
- 🗄️ База данных PostgreSQL для хранения данных пользователей
- 🏥 Health check endpoints для мониторинга

## Архитектура

```
factory/
├── cmd/main.go              # Точка входа
├── internal/
│   ├── config/              # Конфигурация
│   ├── database/            # Подключение к БД
│   ├── handlers/            # HTTP обработчики
│   ├── keycloak/            # Клиент Keycloak
│   ├── logger/              # Структурированное логирование
│   ├── models/              # Модели данных
│   └── services/            # Бизнес-логика
│       ├── plumbus_service.go    # Генерация плюмбусов
│       ├── signature_service.go  # Цифровые подписи
│       └── events_service.go     # NATS события
├── web/
│   ├── templates/           # HTML шаблоны
│   └── static/             # CSS, JS, изображения
├── storage/                 # Хранилище изображений
├── Dockerfile
├── go.mod
└── go.sum
```

## Интеграция с микросервисами

Factory интегрируется со следующими сервисами:

- **plumbus_image_gen** - Генерация изображений плюмбусов
- **vsfi-2025-sig-store** - Создание цифровых подписей
- **vsfi-2025-events-audit** - Аудит событий через NATS
- **PostgreSQL** - Хранение данных пользователей и плюмбусов
- **NATS JetStream** - Событийная шина

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Все сервисы микросервисной архитектуры

### Шаги установки

1. **Запустите всю микросервисную архитектуру**:
   ```bash
   docker-compose up --build
   ```

2. **Сервисы будут доступны**:
   - Factory UI: http://localhost:8082
   - Factory Health: http://localhost:8082/health
   - Plumbus Generator API: http://localhost:8081
   - Sig-Store API: http://localhost:8083
   - Events Audit: http://localhost:8084
   - PostgreSQL: localhost:5432
   - NATS: localhost:4222

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `DATABASE_URL` | Строка подключения к PostgreSQL | `postgres://postgres:accountant@postgres:5432/accountant?sslmode=disable` |
| `KEYCLOAK_URL` | URL Keycloak сервера | `http://localhost:8080` |
| `KEYCLOAK_REALM` | Realm в Keycloak | `master` |
| `KEYCLOAK_CLIENT_ID` | ID клиента в Keycloak | `factory` |
| `KEYCLOAK_CLIENT_SECRET` | Секрет клиента Keycloak | `` |
| `PLUMBUS_SERVICE_URL` | URL сервиса генерации плюмбусов | `http://image-gen:8080` |
| `SIG_STORE_URL` | URL сервиса цифровых подписей | `http://sig-store:8080` |
| `NATS_URL` | URL NATS сервера | `nats://nats:4222` |
| `EVENTS_STREAM` | Имя NATS stream для событий | `events` |
| `EVENTS_SUBJECT` | Subject для событий плюмбусов | `events.plumbus` |
| `SESSION_SECRET` | Ключ для сессий | `your-super-secret-key-here` |
| `PORT` | Порт для запуска сервиса | `8080` |
| `LOG_LEVEL` | Уровень логирования (trace,debug,info,warn,error) | `info` |

## API Endpoints

### Публичные маршруты
- `GET /` - Главная страница
- `GET /health` - Health check endpoint
- `GET /auth/login` - Вход через Keycloak
- `GET /auth/callback` - Callback авторизации
- `GET /auth/logout` - Выход

### Защищенные маршруты (требуют авторизации)
- `GET /dashboard` - Панель управления
- `POST /plumbus/generate` - Создание нового плюмбуса
- `GET /plumbus/status/:id` - Проверка статуса генерации
- `GET /plumbus/image/:id` - Получение изображения плюмбуса
- `GET /plumbus/list` - Список плюмбусов пользователя

## Цифровые подписи

Каждый созданный плюмбус автоматически получает цифровую подпись:

- **Алгоритм**: RSA-SHA256
- **Интеграция**: vsfi-2025-sig-store сервис
- **Отображение**: В UI с возможностью копирования
- **Хранение**: В базе данных вместе с датой подписания

### Процесс подписания

1. Пользователь создает плюмбус
2. Изображение генерируется через plumbus_image_gen
3. Изображение отправляется в sig-store для подписания
4. Подпись сохраняется в базе данных
5. Событие публикуется в NATS

## Событийная архитектура

Factory публикует события в NATS JetStream при создании плюмбусов:

```json
{
  "event_type": "plumbus_created",
  "user_id": "uuid",
  "plumbus_id": "uuid",
  "timestamp": "2025-01-XX:XX:XX.XXXZ",
  "metadata": {
    "name": "My Plumbus",
    "size": "M",
    "color": "blue",
    "shape": "smooth",
    "weight": "medium",
    "wrapping": "standard"
  }
}
```

События обрабатываются сервисом vsfi-2025-events-audit для создания аудит-логов.

## JSON Логирование

Factory использует структурированное логирование с logrus:

### Настройка уровней
- `trace` - Максимальная детализация (включая SQL запросы)
- `debug` - Отладочная информация
- `info` - Основная информация (по умолчанию)
- `warn` - Предупреждения
- `error` - Ошибки

### Примеры логов

```json
{
  "level": "info",
  "msg": "Plumbus generation completed",
  "plumbus_id": "uuid",
  "user_id": "uuid",
  "signature": "RSA-SHA256:...",
  "duration_ms": 1250,
  "timestamp": "2025-01-XX:XX:XX.XXXZ"
}
```

```json
{
  "level": "info",
  "msg": "HTTP Request",
  "method": "POST",
  "path": "/plumbus/generate",
  "status": 200,
  "duration_ms": 45,
  "user_agent": "Mozilla/5.0...",
  "ip": "192.168.1.1",
  "timestamp": "2025-01-XX:XX:XX.XXXZ"
}
```

## Использование

1. **Откройте браузер** и перейдите на http://localhost:8082
2. **Нажмите "Войти"** для авторизации через Keycloak
3. **Заполните форму** создания плюмбуса:
   - Название
   - Размер (nano, XS, S, M, L, XL, XXL)
   - Цвет (12 доступных цветов)
   - Форма (гладкая, угловатая, мульти-угловатая)
   - Вес (сверхлёгкий, лёгкий, средний, тяжёлый)
   - Упаковка (стандартная, подарочная, лимитированная)
4. **Нажмите "Создать плюмбус"** и наблюдайте за прогресс-баром
5. **Просматривайте коллекцию** с цифровыми подписями

## Особенности

- ⚡ **Асинхронная генерация** - плюмбусы создаются в фоновом режиме
- ✍️ **Автоматические подписи** - каждое изображение подписывается цифровой подписью
- 📨 **Событийная архитектура** - все действия публикуются как события
- 📊 **Структурированные логи** - JSON логирование для легкого анализа
- 🔄 **Автообновление статуса** - статус генерации обновляется каждые 5 секунд
- 🎭 **Rick & Morty стилистика** - анимации порталов, частицы, тематические цвета
- 📱 **Адаптивный дизайн** - работает на всех устройствах
- 🛡️ **Безопасность** - JWT токены, защищенные маршруты
- 🏥 **Health checks** - мониторинг состояния сервиса

## Разработка

### Структура базы данных

```sql
-- Пользователи
CREATE TABLE users (
    id UUID PRIMARY KEY,
    keycloak_id VARCHAR UNIQUE NOT NULL,
    username VARCHAR NOT NULL,
    email VARCHAR NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Плюмбусы с цифровыми подписями
CREATE TABLE plumbus (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR NOT NULL,
    size VARCHAR NOT NULL,
    color VARCHAR NOT NULL,
    shape VARCHAR NOT NULL,
    weight VARCHAR NOT NULL,
    wrapping VARCHAR NOT NULL,
    status VARCHAR DEFAULT 'pending',
    image_path VARCHAR,
    signature VARCHAR,          -- Цифровая подпись изображения
    signature_date TIMESTAMP,   -- Дата создания подписи
    error_msg VARCHAR,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Локальная разработка

```bash
# Установка зависимостей
go mod tidy

# Запуск в режиме разработки с debug логами
LOG_LEVEL=debug go run cmd/main.go

# Сборка
go build -o factory cmd/main.go
```

### Тестирование

#### Запуск unit-тестов

Для запуска всех unit-тестов сервиса:

```bash
cd factory
make test              # Быстрый запуск всех тестов
make test-verbose      # Подробный вывод тестов
```

Для запуска тестов с покрытием кода:

```bash
make test-coverage     # Генерирует coverage.html с отчетом
```

Для запуска тестов конкретного модуля:

```bash
make test-models       # Тесты моделей данных
make test-config       # Тесты конфигурации  
make test-services     # Тесты сервисов
```

Полная справка по командам:

```bash
make help             # Показать все доступные команды
```

#### Тестовые зависимости

Unit-тесты используют:
- **SQLite** - in-memory база данных для тестирования UserService
- **HTTP моки** - для тестирования PlumbusService и SignatureService  
- **NATS моки** - для тестирования EventsService

Для установки тестовых зависимостей:

```bash
go mod download
go get gorm.io/driver/sqlite
```

#### Структура тестов

```
internal/
├── models/
│   ├── models.go
│   └── models_test.go      # Тесты структур данных
├── config/
│   ├── config.go
│   └── config_test.go      # Тесты конфигурации
├── services/
│   ├── plumbus.go
│   ├── plumbus_test.go     # Тесты генерации плюмбусов
│   ├── signature.go
│   ├── signature_test.go   # Тесты цифровых подписей
│   ├── events.go
│   ├── events_test.go      # Тесты событий NATS
│   ├── user.go
│   └── user_test.go        # Тесты работы с пользователями
└── testutils/
    └── testutils.go        # Общие тестовые утилиты
```

#### Покрытие тестами

Тесты покрывают:
- ✅ Модели данных и их валидацию
- ✅ Загрузку конфигурации из переменных окружения
- ✅ HTTP взаимодействие с внешними сервисами
- ✅ Работу с базой данных (CRUD операции)
- ✅ Публикацию событий в NATS
- ✅ Обработку ошибок и edge cases

### Зависимости

Основные Go модули:
- `github.com/gin-gonic/gin` - HTTP фреймворк
- `gorm.io/gorm` - ORM для работы с БД
- `github.com/sirupsen/logrus` - Структурированное логирование
- `github.com/nats-io/nats.go` - NATS клиент
- `github.com/google/uuid` - Генерация UUID
- `github.com/joho/godotenv` - Переменные окружения

## Интеграция с Keycloak

Для корректной работы авторизации настройте клиента в Keycloak:

1. Авторизуйтесь в Keycloak. admin:admin
2. Создайте realm (или используйте существующий)
![realm](./img/realm1.png)

![realm](./img/realm.png)
3. Создайте клиента с ID `factory`
![client](./img/client.png)

![client](./img/client1.png)

![client](./img/client2.png)

![client](./img/client3.png)

4. Скопируйте Client Secret в переменную окружения
![secret](./img/secret.png)

5. Создайте пользователя

![usr](./img/usr.png)

![usr](./img/usr1.png)

![usr](./img/usr2.png)

![usr](./img/usr3.png)

### Просмотр логов
```bash
# Просмотр логов всех сервисов
docker-compose logs -f

# Логи только factory в JSON формате
docker-compose logs -f factory | jq '.'

# Фильтрация логов по уровню
docker-compose logs -f factory | jq 'select(.level == "error")'
```

### Health checks
```bash
# Проверка состояния factory
curl http://localhost:8082/health

# Проверка всех сервисов
curl http://localhost:8082/health && \
curl http://localhost:8081/health && \
curl http://localhost:8083/health && \
curl http://localhost:8084/health
```

### NATS мониторинг
```bash
# Проверка NATS streams
docker exec -it pepyaka-nats-1 nats stream ls

# Просмотр событий
docker exec -it pepyaka-nats-1 nats stream view events
```

## Troubleshooting

### Проблемы с подключением к БД
- Убедитесь что PostgreSQL запущен
- Проверьте правильность DATABASE_URL
- Проверьте логи: `docker-compose logs postgres`

### Проблемы с авторизацией
- Проверьте настройки Keycloak клиента
- Убедитесь что KEYCLOAK_* переменные правильно настроены
- Проверьте redirect URIs в Keycloak

### Проблемы с генерацией плюмбусов
- Проверьте доступность plumbus_image_gen: `curl http://localhost:8081/health`
- Убедитесь что PLUMBUS_SERVICE_URL правильно настроен

### Проблемы с подписями
- Проверьте доступность sig-store: `curl http://localhost:8083/health`
- Убедитесь что SIG_STORE_URL правильно настроен
- Проверьте логи sig-store: `docker-compose logs sig-store`

### Проблемы с событиями
- Проверьте статус NATS: `docker-compose logs nats`
- Убедитесь что NATS_URL правильно настроен
- Проверьте создание stream: `docker exec -it pepyaka-nats-1 nats stream ls`

### Анализ логов
```bash
# Поиск ошибок в логах
docker-compose logs factory | jq 'select(.level == "error")'

# Анализ производительности
docker-compose logs factory | jq 'select(.duration_ms > 1000)'

# Мониторинг событий NATS
docker-compose logs factory | jq 'select(.msg | contains("NATS"))'
```

## Лицензия

© 2025 Rick Sanchez Enterprises. Все права защищены в бесконечном количестве измерений.

---

*Wubba lubba dub dub!* 🛸 