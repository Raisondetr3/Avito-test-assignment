# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы.

## Запуск

Запустить сервис и базу данных:
```bash
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`

Остановить сервис:
```bash
docker-compose down
```

## Makefile команды

```bash
make build              # Собрать Docker образ
make up                 # Запустить сервис
make down               # Остановить сервис
make integration-test   # Запустить интеграционные тесты
make lint               # Запустить линтер
make clean              # Очистить Docker ресурсы
```

## API

API описан в `openapi.yml`

## Результаты нагрузочного тестирования

**Условия:** 5 RPS, до 200 пользователей, 20 команд

**Результаты:**
- P50 latency: 7.92 ms
- P95 latency: 18.53 ms (цель: <300 ms)
- P99 latency: 45.38 ms
- Success rate: 99.33% (цель: >99.9%)
- Bulk deactivation: 6-8 ms (цель: <100 ms)


## Дополнительные возможности

Реализованы все опциональные задания:
- ✓ Эндпоинт статистики (`GET /stats`)
- ✓ Нагрузочное тестирование
- ✓ Массовая деактивация пользователей (`POST /team/deactivateUsers`)
- ✓ Интеграционное тестирование
- ✓ Конфигурация линтера

## Стек

- Go 1.23
- PostgreSQL 16
- Docker & Docker Compose
