package utils

import (
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

const (
	limiterShardCount      = 64
	ipIdleTTL              = 10 * time.Minute
	lastSeenUpdateInterval = time.Minute
	cleanupInterval        = 5 * time.Minute
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

func (e *entry) touch(now int64) {
	last := e.lastSeen.Load()
	if now-last < int64(lastSeenUpdateInterval) {
		return
	}

	e.lastSeen.CompareAndSwap(last, now)
}

type limiterShard struct {
	mu  sync.RWMutex
	ips map[string]*entry
}

type IPRateLimiter struct {
	shards [limiterShardCount]limiterShard
	r      rate.Limit
	b      int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		r: r,
		b: b,
	}

	for shard := range i.shards {
		i.shards[shard].ips = make(map[string]*entry)
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UnixNano()

			for shard := range i.shards {
				s := &i.shards[shard]
				s.mu.Lock()
				for ip, e := range s.ips {
					if now-e.lastSeen.Load() > int64(ipIdleTTL) {
						delete(s.ips, ip)
					}
				}
				s.mu.Unlock()
			}
		}
	}()

	return i
}

func limiterShardIndex(ip string) uint64 {
	const (
		offsetBasis uint64 = 14695981039346656037
		prime       uint64 = 1099511628211
	)

	hash := offsetBasis
	for index := 0; index < len(ip); index++ {
		hash ^= uint64(ip[index])
		hash *= prime
	}

	return hash % limiterShardCount
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	shard := &i.shards[limiterShardIndex(ip)]
	now := time.Now().UnixNano()

	shard.mu.RLock()
	e, exists := shard.ips[ip]
	shard.mu.RUnlock()

	if exists {
		e.touch(now)
		return e.limiter
	}

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if e, exists = shard.ips[ip]; exists {
		e.touch(now)
		return e.limiter
	}

	e = &entry{
		limiter: rate.NewLimiter(i.r, i.b),
	}
	e.lastSeen.Store(now)
	shard.ips[ip] = e

	return e.limiter
}

func getEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}

	return parsed
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
