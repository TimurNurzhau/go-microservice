package utils

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
	// Периодическая очистка старых IP, чтобы map не рос бесконечно
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			i.mu.Lock()
			i.ips = make(map[string]*rate.Limiter)
			i.mu.Unlock()
		}
	}()
	return i
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if exists {
		return limiter
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	// Повторная проверка после захвата блокировки (double-check)
	if limiter, exists = i.ips[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter
	return limiter
}

// Лимитер на 1000 rps / burst 5000 согласно ТЗ (per-IP).
var ipLimiter = NewIPRateLimiter(1000, 5000)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		limiter := ipLimiter.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}