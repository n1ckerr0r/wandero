# Backend

Go API для Wandero. Контракт находится в `api/openapi.yaml`, generated routing — в `internal/openapi`, реализация — в `internal/api`.

```bash
GOCACHE=/tmp/chronograph-go-cache go test -race ./...
GOCACHE=/tmp/chronograph-go-cache go vet ./...
go run ./cmd/server
```

## Миграции

`cmd/migrate` читает встроенные `migrations/*.sql`, берёт PostgreSQL advisory lock и применяет каждый новый файл в транзакции. Имя записывается в `schema_migrations` только после успешного commit.

```bash
go run ./cmd/migrate
```

Не изменяйте уже применённые SQL-файлы; добавляйте новый файл с возрастающим номером. Seed-данные являются частью текущей истории миграций.

## HTTP hardening

Сервер настроен с read/write/idle timeout, panic recovery, ограничением JSON body в 1 MiB, запретом неизвестных JSON-полей, безопасными публичными ошибками и graceful shutdown. `CORS_ALLOWED_ORIGINS` содержит разрешённые origins через запятую. Если `API_WRITE_KEY` задан, mutation-методы требуют `Authorization: Bearer ...`.

Полная таблица env и команды запуска приведены в корневом README.
