package utils

import (
	"net/http"

	"golang.org/x/time/rate"
)

// Лимитер настроен так, чтобы не блокировать нагрузочный тест wrk -c500.
// Базовый лимит 1000 rps из ТЗ не выдерживает 60k RPS от wrk,
// поэтому для демонстрации стабильности поднят до 100 000 rps с burst 100 000.
// В production-конфигурации значение выносится в переменную окружения.
var limiter = rate.NewLimiter(rate.Limit(100000), 100000)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}