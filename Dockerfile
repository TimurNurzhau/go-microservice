# Этап сборки
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и собираем бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный этап (минимальный образ)
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]