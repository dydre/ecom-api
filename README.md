# ecom-api

REST API для e-commerce на Go + PostgreSQL.

**Стек:** chi, pgx/v5, sqlc, cleanenv, Docker

---

## Быстрый старт

```bash
# 1. Запустить PostgreSQL
docker compose up -d

# 2. Применить миграцию
docker exec -i ecom-postgres psql -U postgres -d ecom \
  < internal/adapters/postgresql/migrations/00001_create_products.sql

# 3. Запустить сервер
go run ./cmd/main.go
```

Сервер запустится на `http://localhost:8080`.

---

## Конфиг

Файл: `config/local.yaml`. Можно переопределить через переменные окружения.

| Переменная   | Default              | Описание               |
|--------------|----------------------|------------------------|
| `CONFIG_PATH`| `config/local.yaml`  | Путь к конфиг-файлу    |
| `ENV`        | `local`              | `local` / `dev` / `prod` |
| `DB_DSN`     | —                    | PostgreSQL DSN (обязательно) |
| `HTTP_PORT`  | `8080`               | Порт сервера           |

---

## Endpoints

| Method | Path        | Описание         |
|--------|-------------|------------------|
| GET    | `/health`   | Health check     |
| GET    | `/products` | Список продуктов |

---

## Работа с БД (sqlc)

Новый запрос:
1. Написать SQL в `internal/adapters/postgresql/sqlc/queries.sql`
2. Запустить `sqlc generate`

Новая миграция:
- Создать файл `internal/adapters/postgresql/migrations/00002_<name>.sql`

