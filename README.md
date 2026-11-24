# Go Microservice — CRUD API

HTTP-сервис на Go для управления пользователями. Реализованы CRUD API, ограничение частоты запросов, Prometheus-метрики, audit logging и запуск в Docker Compose.

## Технологии

- Go 1.22+
- gorilla/mux — HTTP-роутинг
- golang.org/x/time/rate — rate limiting
- prometheus/client_golang — метрики
- minio-go/v7 — S3 API client
- Docker Compose
- Silo — локальное S3-совместимое хранилище для разработки

> В текущей конфигурации Compose используется Silo, а не MinIO. Если для сдачи требуется именно MinIO, замените образ хранилища на доступный образ MinIO и проверьте запуск на чистой машине.

## Запуск

```powershell
docker compose up -d --build
docker compose ps
docker compose logs app --tail=100
```

Проверить работоспособность:

```powershell
curl.exe http://localhost:8080/health
curl.exe http://localhost:8080/api/users
curl.exe http://localhost:8080/metrics
```

Остановить контейнеры, сохранив volume:

```powershell
docker compose down
```

## API

| Метод | URL | Описание |
|---|---|---|
| GET | `/api/users` | Получить всех пользователей |
| GET | `/api/users/{id}` | Получить пользователя по ID |
| POST | `/api/users` | Создать пользователя |
| PUT | `/api/users/{id}` | Обновить пользователя |
| DELETE | `/api/users/{id}` | Удалить пользователя |
| GET | `/metrics` | Получить метрики Prometheus |
| GET | `/health` | Проверить состояние сервиса |

Пример создания пользователя:

```powershell
curl.exe -X POST http://localhost:8080/api/users `
  -H "Content-Type: application/json" `
  -d '{"name":"Ivan","email":"ivan@example.com"}'
```

## Метрики Prometheus

Эндпоинт метрик: `http://localhost:8080/metrics`.

Экспортируются:

- `http_requests_total` — количество запросов;
- `http_request_duration_seconds` — длительность запросов;
- `http_errors_total` — ответы с кодами 4xx и 5xx.

RPS можно получить в Prometheus запросом:

```promql
rate(http_requests_total[1m])
```

Средняя latency за период:

```promql
rate(http_request_duration_seconds_sum[1m])
/
rate(http_request_duration_seconds_count[1m])
```

## Rate limiting

По умолчанию в Compose настроены:

- `RATE_LIMIT=1000` запросов в секунду на IP;
- `RATE_BURST=5000` запросов burst.

При превышении лимита API возвращает HTTP 429. Значения можно переопределить переменными окружения.

## Нагрузочное тестирование с wrk

Запустить тест из Docker-сети проекта:

```powershell
docker run --rm --network microservice-net williamyeh/wrk `
  -t12 -c500 -d60s `
  http://app:8080/api/users
```

Сохраните полный вывод теста, включая `Requests/sec`, `Latency` и строки `Socket errors`. Указывайте в отчёте только фактически полученные результаты. Результаты зависят от CPU, памяти, Docker Desktop/WSL2 и настроек rate limiter.

## Хранение данных

CRUD-данные пользователей хранятся в памяти процесса, поэтому удаляются при перезапуске приложения. S3-клиент при запуске проверяет подключение к хранилищу и инициализирует бакет. Пользовательские записи в S3 в текущей реализации не сохраняются.