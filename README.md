# Wandero

Wandero — приложение для городских прогулок по Санкт-Петербургу. React/Vite-клиент показывает места и подборки маршрутов на MapLibre-карте; нативный Android-клиент ведёт по прогулочным сценариям. Go API отдаёт JSON/GeoJSON, данные хранятся в PostgreSQL/PostGIS.

## Быстрый старт

Требования: Docker с Compose v2.20+ и `make`.

```bash
make up
```

После запуска:

- приложение: <http://localhost:3000>;
- API healthcheck: <http://localhost:8080/health>;
- Swagger UI: `docker compose --profile docs up -d swagger-ui`, затем <http://localhost:8081>;
- PostgreSQL доступен только на loopback-порту `55432`.

Первый запуск применяет SQL-миграции отдельным одноразовым контейнером `migrate`. Повторные запуски применяют только новые файлы. Старые базы, созданные прежним `docker-entrypoint-initdb.d`-механизмом, автоматически baseline-ятся без повторного запуска seed-миграций.

Полезные команды:

```bash
make help
make check          # формат, линты, тесты, сборка и Compose validation
make logs
make down
make migrate
```

## Локальная разработка

```bash
make install

# terminal 1
cd backend
GOCACHE=/tmp/chronograph-go-cache go run ./cmd/server

# terminal 2
cd frontend
npm run dev
```

Vite проксирует `/api` и `/health` на `localhost:8080`. Production frontend использует Nginx reverse proxy, поэтому браузеру не нужен отдельный адрес API и CORS.

Основные каталоги:

```text
backend/api/openapi.yaml       источник истины API
frontend/src/api-schema.d.ts  сгенерированные TypeScript-типы API
backend/cmd/server             HTTP-сервер
backend/cmd/migrate            мигратор
backend/internal/api           обработчики и middleware
backend/migrations             схема и seed-данные
frontend/src                   React-приложение
android/app                    нативное Android-приложение (Compose)
docs/openapi.yaml              синхронизированная копия для документации
.github/workflows/ci.yml       CI
```

После изменения OpenAPI:

```bash
make generate
make check
```

`make generate` обновляет Go server bindings, копию спецификации в `docs/` и TypeScript-схему фронтенда. Ручной API-клиент проверяет пути по сгенерированному типу `paths`; CI дополнительно сравнивает результат генерации с закоммиченной схемой.

Сверка 27 операций с UI и список недостающих route API — в [docs/api-coverage.md](docs/api-coverage.md).

## Android

Мобильная версия Wandero переносит сценарии `M10–M16` из Figma в нативный Jetpack Compose: карта и места, карточка места, сборка прогулки, готовый маршрут, навигация и каталог популярных маршрутов.

```bash
GRADLE_USER_HOME=/tmp/wandero-gradle gradle -p android assembleDebug
```

Подробности — в [android/README.md](android/README.md).

## Проверки качества

`make check` выполняет:

- `gofmt` check, `go vet`, `go test -race -cover`;
- Vitest-тесты API-клиента и пользовательских действий боковой панели;
- проверку синхронности OpenAPI-контрактов;
- ESLint с type-aware правилами, Prettier и TypeScript typecheck;
- production-сборки backend/frontend;
- валидацию dev и production Compose-конфигураций.

Интеграционные Go-тесты включаются переменной `TEST_DATABASE_URL`:

```bash
docker compose up -d postgres migrate
TEST_DATABASE_URL=postgres://chronograph:chronograph@localhost:55432/chronograph \
  make test-backend
```

Для API-интеграционных тестов тестовая база должна быть отдельной: тесты сбрасывают схему и применяют миграции заново.

## Production deployment

Проект содержит production overlay для одного сервера. TLS должен завершаться во внешнем reverse proxy/load balancer; наружу публикуется только frontend. Backend и PostgreSQL в базовом Compose привязаны к `127.0.0.1`.

```bash
cp .env.example .env
# Замените DB_PASSWORD и API_WRITE_KEY на независимые длинные случайные значения.
docker compose -f docker-compose.yml -f docker-compose.prod.yml config
make prod-up
curl --fail http://localhost:8080/health
```

В production overlay непустые `DB_PASSWORD` и `API_WRITE_KEY` обязательны. Все `POST`, `PUT` и `DELETE` запросы требуют заголовок:

```text
Authorization: Bearer <API_WRITE_KEY>
```

Читайте полный checklist в [docs/deployment.md](docs/deployment.md).

## Конфигурация backend

| Переменная | Default | Назначение |
| --- | --- | --- |
| `SERVER_ADDR` | `:8080` | адрес HTTP-сервера |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | dev defaults | PostgreSQL |
| `DB_SSLMODE` | `disable` | режим TLS pgx |
| `CORS_ALLOWED_ORIGINS` | пусто | список origins через запятую |
| `API_WRITE_KEY` | пусто | Bearer key для mutation API; обязателен в prod overlay |
| `HTTP_READ_TIMEOUT` | `10s` | read/read-header timeout |
| `HTTP_WRITE_TIMEOUT` | `30s` | write timeout |
| `HTTP_IDLE_TIMEOUT` | `60s` | keep-alive timeout |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | graceful shutdown timeout |

## Важные ограничения

- Это публичный read API со статическим ключом для записей, не полноценная пользовательская RBAC-система.
- Административного UI и загрузки файлов нет; media содержит внешние HTTP(S)-URL.
- Пагинация возвращает массив без total/meta.
- Карта и seed-изображения используют внешние OpenStreetMap/Wikimedia-ресурсы; для большой нагрузки нужен собственный или коммерческий tile provider с подходящим SLA.
- Резервное копирование PostgreSQL и TLS выполняются инфраструктурой деплоя, а не приложением.
