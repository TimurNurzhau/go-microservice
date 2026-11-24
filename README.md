# Go Microservice (High-Load CRUD API)

Микросервис на Go для управления пользователями с поддержкой высоких нагрузок.

## Стек
- Go 1.26+
- gorilla/mux (роутинг)
- prometheus/client_golang (метрики)
- golang.org/x/time/rate (rate limiting)
- minio-go/v7 (S3-хранилище)

## Запуск
```bash
docker compose up -d --build

Нагрузочное тестирование
docker run --rm --network go-microservice_microservice-net williamyeh/wrk -t12 -c500 -d60s http://app:8080/api/users

Эндпоинты
GET /api/users

GET /api/users/{id}

POST /api/users

PUT /api/users/{id}

DELETE /api/users/{id}

GET /metrics

GET /health

### Шаг 3. Делаем `git push`

Выполните последовательно:

```powershell
cd C:\projects\go-microservice
git add .
git commit -m "Final version: high-load microservice with rate limiter, metrics, and MinIO integration"
git push origin main