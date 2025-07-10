# Plumbus Factory 🛸

Rick & Morty стилизованный веб-сервис для создания плюмбусов с интеграцией Keycloak для авторизации.

## Описание

Factory - это Go веб-сервис, который предоставляет пользовательский интерфейс для создания плюмбусов, используя существующий сервис `plumbus_image_gen`. Проект включает:

- 🔐 Авторизация через Keycloak
- 🎨 Современный UI в стилистике Rick & Morty
- 📊 Прогресс-бар генерации
- 💾 Персональное хранилище изображений
- 🗄️ База данных PostgreSQL для хранения данных пользователей

## Архитектура

```
factory/
├── cmd/main.go              # Точка входа
├── internal/
│   ├── config/              # Конфигурация
│   ├── database/            # Подключение к БД
│   ├── handlers/            # HTTP обработчики
│   ├── keycloak/            # Клиент Keycloak
│   ├── models/              # Модели данных
│   └── services/            # Бизнес-логика
├── web/
│   ├── templates/           # HTML шаблоны
│   └── static/             # CSS, JS, изображения
├── storage/                 # Хранилище изображений
├── Dockerfile
├── go.mod
└── go.sum
```

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- PostgreSQL (через docker-compose)
- Сервис plumbus_image_gen (уже включен)

### Шаги установки

1. **Клонируйте репозиторий** (если ещё не сделано)

2. **Настройте переменные окружения**:
   ```bash
   cp factory/.env.example factory/.env
   # Отредактируйте .env файл под ваши настройки
   ```

3. **Запустите все сервисы**:
   ```bash
   docker-compose up --build
   ```

4. **Сервис будет доступен**:
   - Factory UI: http://localhost:8082
   - Plumbus Generator API: http://localhost:8081
   - PostgreSQL: localhost:5432

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `DATABASE_URL` | Строка подключения к PostgreSQL | `postgres://postgres:accountant@postgres:5432/accountant?sslmode=disable` |
| `KEYCLOAK_URL` | URL Keycloak сервера | `http://localhost:8080` |
| `KEYCLOAK_REALM` | Realm в Keycloak | `master` |
| `KEYCLOAK_CLIENT_ID` | ID клиента в Keycloak | `factory` |
| `KEYCLOAK_CLIENT_SECRET` | Секрет клиента Keycloak | `` |
| `PLUMBUS_SERVICE_URL` | URL сервиса генерации плюмбусов | `http://image-gen:8080` |
| `SESSION_SECRET` | Ключ для сессий | `your-super-secret-key-here` |
| `PORT` | Порт для запуска сервиса | `8080` |

## API Endpoints

### Публичные маршруты
- `GET /` - Главная страница
- `GET /auth/login` - Вход через Keycloak
- `GET /auth/callback` - Callback авторизации
- `GET /auth/logout` - Выход

### Защищенные маршруты (требуют авторизации)
- `GET /dashboard` - Панель управления
- `POST /plumbus/generate` - Создание нового плюмбуса
- `GET /plumbus/status/:id` - Проверка статуса генерации
- `GET /plumbus/image/:id` - Получение изображения плюмбуса
- `GET /plumbus/list` - Список плюмбусов пользователя

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
5. **Просматривайте коллекцию** созданных плюмбусов

## Особенности

- ⚡ **Асинхронная генерация** - плюмбусы создаются в фоновом режиме
- 🔄 **Автообновление статуса** - статус генерации обновляется каждые 5 секунд
- 🎭 **Rick & Morty стилистика** - анимации порталов, частицы, тематические цвета
- 📱 **Адаптивный дизайн** - работает на всех устройствах
- 🛡️ **Безопасность** - JWT токены, защищенные маршруты

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

-- Плюмбусы
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
    error_msg VARCHAR,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Локальная разработка

```bash
# Установка зависимостей
go mod tidy

# Запуск в режиме разработки
go run cmd/main.go

# Сборка
go build -o factory cmd/main.go
```

## Интеграция с Keycloak

Для корректной работы авторизации настройте клиента в Keycloak:

1. Создайте realm (или используйте существующий)
2. Создайте клиента с ID `factory`
3. Скопируйте Client Secret в переменную окружения

## Мониторинг и логи

```bash
# Просмотр логов всех сервисов
docker-compose logs -f

# Логи только factory
docker-compose logs -f factory

# Логи только plumbus_image_gen
docker-compose logs -f image-gen
```

## Troubleshooting

### Проблемы с подключением к БД
- Убедитесь что PostgreSQL запущен
- Проверьте правильность DATABASE_URL

### Проблемы с авторизацией
- Проверьте настройки Keycloak клиента
- Убедитесь что KEYCLOAK_* переменные правильно настроены

### Проблемы с генерацией плюмбусов
- Проверьте что сервис plumbus_image_gen доступен
- Убедитесь что PLUMBUS_SERVICE_URL правильно настроен

## Лицензия

© 2025 Rick Sanchez Enterprises. Все права защищены в бесконечном количестве измерений.

---

*Wubba lubba dub dub!* 🛸 