# AutoFluent

CRM-система управления автосервисом. REST API на Go с ETL-конвейером для формирования отчётности.

## Стек

- **Go 1.21** — язык разработки
- **chi v5** — HTTP-маршрутизатор
- **PostgreSQL 16** — база данных
- **pgx v5 + sqlx** — драйвер и утилиты для работы с БД
- **golang-jwt/jwt v5** — JWT-аутентификация
- **go.uber.org/zap** — структурированное логирование
- **godotenv** — конфигурация через переменные окружения
- **Docker + Docker Compose** — контейнеризация

## Структура проекта

```
.
├── cmd/
│   ├── api/main.go          # Точка входа HTTP-сервера
│   └── etl/main.go          # Точка входа ETL-конвейера
├── internal/
│   ├── api/
│   │   ├── http/            # HTTP-хендлеры (client.go, order.go, handler.go)
│   │   └── mid/             # Middleware: JWT-auth, logger, recover
│   ├── domain/              # Бизнес-сущности и доменные ошибки
│   ├── etl/                 # ETL: extractor → transformer → loader → pipeline
│   ├── repository/
│   │   ├── postgres/        # PostgreSQL-реализация репозиториев
│   │   └── store.go         # Интерфейс хранилища
│   └── service/             # Бизнес-логика: client, order, stock
├── migrations/              # SQL-миграции схемы БД
├── pkg/
│   ├── config/              # Загрузка конфигурации из env
│   └── logger/              # Инициализация zap-логгера
└── reports/                 # JSON-отчёты, генерируемые ETL-конвейером
```

## Быстрый старт

### Через Docker Compose

```bash
cp .env.example .env   # заполнить переменные окружения
docker compose up -d
```

### Локально

```bash
# Поднять только PostgreSQL
docker compose up -d postgres

# Применить миграции
make migrate

# Запустить API-сервер
make run
# или
go run ./cmd/api
```

### Сборка бинарников

```bash
make build
# → bin/api, bin/etl
```

## Конфигурация

Переменные окружения (файл `.env`):

| Переменная      | Описание                        | Пример                                      |
|-----------------|---------------------------------|---------------------------------------------|
| `DATABASE_URL`  | DSN для подключения к PostgreSQL | `postgres://user:pass@localhost:5432/autofluent` |
| `JWT_SECRET`    | Секрет для подписи JWT-токенов  | `supersecret`                               |
| `HTTP_ADDR`     | Адрес и порт HTTP-сервера       | `:8080`                                     |
| `LOG_LEVEL`     | Уровень логирования             | `info`                                      |

## API

Все маршруты — под префиксом `/api/v1`. Аутентификация — JWT в заголовке `Authorization: Bearer <token>`.

### Клиенты `/clients`

| Метод    | Путь               | Описание                    |
|----------|--------------------|-----------------------------|
| `POST`   | `/clients`         | Создать клиента             |
| `GET`    | `/clients`         | Список клиентов             |
| `GET`    | `/clients/{id}`    | Получить клиента по ID      |
| `PUT`    | `/clients/{id}`    | Обновить данные клиента     |
| `DELETE` | `/clients/{id}`    | Удалить клиента             |

### Транспортные средства `/vehicles`

| Метод    | Путь                          | Описание                       |
|----------|-------------------------------|--------------------------------|
| `POST`   | `/clients/{id}/vehicles`      | Добавить автомобиль клиенту    |
| `GET`    | `/clients/{id}/vehicles`      | Список автомобилей клиента     |
| `GET`    | `/vehicles/{id}`              | Получить автомобиль по ID      |
| `PUT`    | `/vehicles/{id}`              | Обновить данные автомобиля     |
| `DELETE` | `/vehicles/{id}`              | Удалить автомобиль             |

### Заказы `/orders`

| Метод    | Путь                          | Описание                       |
|----------|-------------------------------|--------------------------------|
| `POST`   | `/orders`                     | Создать заказ                  |
| `GET`    | `/orders`                     | Список заказов (фильтры: `status`, `vehicle_id`) |
| `GET`    | `/orders/{id}`                | Получить заказ по ID           |
| `POST`   | `/orders/{id}/transition`     | Изменить статус заказа         |
| `DELETE` | `/orders/{id}`                | Удалить заказ                  |

Допустимые переходы статусов: `accepted` → `in_progress` → `done` → `issued`.

### Работы и запчасти по заказу

| Метод    | Путь                                  | Описание                        |
|----------|---------------------------------------|---------------------------------|
| `POST`   | `/orders/{id}/services`               | Добавить работу в заказ         |
| `GET`    | `/orders/{id}/services`               | Список работ по заказу          |
| `DELETE` | `/orders/{id}/services/{lineId}`      | Удалить работу из заказа        |
| `POST`   | `/orders/{id}/parts`                  | Добавить запчасть в заказ       |
| `GET`    | `/orders/{id}/parts`                  | Список запчастей по заказу      |
| `DELETE` | `/orders/{id}/parts/{lineId}`         | Удалить запчасть из заказа      |
| `POST`   | `/orders/{id}/payments`               | Добавить платёж по заказу       |
| `GET`    | `/orders/{id}/payments`               | Список платежей по заказу       |

### Каталоги и склад

| Метод  | Путь                          | Описание                        |
|--------|-------------------------------|---------------------------------|
| `POST` | `/services`                   | Добавить услугу в каталог       |
| `GET`  | `/services`                   | Список услуг                    |
| `POST` | `/parts`                      | Добавить запчасть в каталог     |
| `GET`  | `/parts`                      | Список запчастей                |
| `GET`  | `/parts/{id}`                 | Запчасть по ID                  |
| `GET`  | `/stock`                      | Все складские позиции           |
| `GET`  | `/stock/{partId}`             | Остаток по запчасти             |
| `POST` | `/stock/{partId}/replenish`   | Пополнить склад                 |

## ETL-конвейер

Генерирует JSON-отчёт по всем заказам с расчётом долга и статуса оплаты.

```bash
# Запуск локально
make etl
# или
go run ./cmd/etl

# Запуск в Docker
docker compose run --rm etl
```

Отчёт сохраняется в `reports/report_<timestamp>.json`:

```json
[
  {
    "OrderID":     "ORD-0042",
    "ClientName":  "Петров Иван Сергеевич",
    "Phone":       "+79171234567",
    "Vehicle":     "Toyota Camry (А123ВС777)",
    "Status":      "done",
    "Complaint":   "Замена масла",
    "TotalAmount": 3500.00,
    "PaidTotal":   3500.00,
    "Debt":        0,
    "IsPaid":      true
  }
]
```

## Миграции

```bash
# Применить все миграции
make migrate

# Или вручную
psql $DATABASE_URL -f migrations/001_init_schema.sql
psql $DATABASE_URL -f migrations/002_add_stock_idx.sql
```

## Makefile

```bash
make build    # собрать bin/api и bin/etl
make run      # запустить API-сервер
make etl      # запустить ETL-конвейер
make migrate  # применить миграции
make lint     # линтер
make test     # тесты
```
