# Deployment checklist

## Перед первым запуском

- Установите Docker Engine/Compose и настройте внешний TLS reverse proxy.
- Создайте `.env` из `.env.example`; сгенерируйте независимые случайные `DB_PASSWORD` и `API_WRITE_KEY`.
- Ограничьте firewall: публичными должны быть только 80/443 reverse proxy. Порты backend/PostgreSQL по умолчанию слушают loopback.
- Проверьте конфигурацию: `make compose-config`.
- Запустите `make prod-up` и проверьте `docker compose ps`, `/health`, логи мигратора и frontend.

## Обновление

1. Сделайте backup PostgreSQL и проверьте его восстановление.
2. Соберите immutable образы из конкретного commit/tag.
3. Выполните `docker compose ... run --rm migrate`.
4. Обновите backend/frontend и проверьте healthcheck.
5. Для rollback приложения используйте предыдущие образы. SQL-миграции должны быть backward-compatible; автоматического down-migrate намеренно нет.

## Эксплуатация

- Настройте ежедневный `pg_dump`/snapshot, retention и регулярный restore drill.
- Собирайте container logs, метрики CPU/RAM/disk и алерты на `/health`/restart count.
- Контролируйте свободное место PostgreSQL volume и срок действия TLS-сертификатов.
- Ротируйте write key и пароль БД; не коммитьте `.env`.
- За reverse proxy настройте request/body limits и rate limiting для API.
- Используйте выделенный tile provider при заметном трафике: публичный OSM tile endpoint не является production SLA.
