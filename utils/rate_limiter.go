package utils

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

type IPRateLimiter struct {
	ips map[string]*entry
	mu  sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*entry),
		r:   r,
		b:   b,
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UnixNano()

			i.mu.Lock()
			for ip, e := range i.ips {
				if now-e.lastSeen.Load() > int64(10*time.Minute) {
					delete(i.ips, ip)
				}
			}
			i.mu.Unlock()
		}
	}()

	return i
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	e, exists := i.ips[ip]
	if exists {
		e.lastSeen.Store(time.Now().UnixNano())
	}
	i.mu.RUnlock()

	if exists {
		return e.limiter
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if e, exists = i.ips[ip]; exists {
		e.lastSeen.Store(time.Now().UnixNano())
		return e.limiter
	}

	e = &entry{limiter: rate.NewLimiter(i.r, i.b)}
	e.lastSeen.Store(time.Now().UnixNano())
	i.ips[ip] = e

	return e.limiter
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

var ipLimiter = NewIPRateLimiter(
	rate.Limit(getEnvFloat("RATE_LIMIT", 1000)),
	getEnvInt("RATE_BURST", 5000),
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		if !ipLimiter.GetLimiter(ip).Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
