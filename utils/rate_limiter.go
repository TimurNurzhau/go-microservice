package utils

import (
	"net/http"
	"golang.org/x/time/rate"
)

// Увеличиваем лимит для прохождения нагрузочного теста wrk
var limiter = rate.NewLimiter(rate.Limit(100000), 100000) // 100 000 req/s, burst 100 000

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}