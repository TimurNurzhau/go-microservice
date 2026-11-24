package utils

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
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
		for {
			time.Sleep(5 * time.Minute)
			i.mu.Lock()
			for ip, e := range i.ips {
				if time.Since(e.lastSeen) > 10*time.Minute {
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
	i.mu.RUnlock()

	if exists {
		i.mu.Lock()
		e.lastSeen = time.Now()
		i.mu.Unlock()
		return e.limiter
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if e, exists = i.ips[ip]; exists {
		e.lastSeen = time.Now()
		return e.limiter
	}

	e = &entry{
		limiter:  rate.NewLimiter(i.r, i.b),
		lastSeen: time.Now(),
	}
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
