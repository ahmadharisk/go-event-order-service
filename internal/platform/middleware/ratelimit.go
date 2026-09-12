package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimit token bucket per IP sederhana (in-memory untuk dev; prod pindah ke Redis INCR+EXPIRE)
type limiter struct {
	tokens float64
	last   time.Time
}

var (
	mu       sync.Mutex
	buckets  = map[string]*limiter{}
	rate     = 100.0 // req per 60s
	interval = time.Minute
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		mu.Lock()
		b, ok := buckets[ip]
		if !ok {
			b = &limiter{tokens: rate, last: time.Now()}
			buckets[ip] = b
		}
		elapsed := time.Since(b.last).Seconds()
		b.tokens += elapsed * (rate / interval.Seconds())
		if b.tokens > rate {
			b.tokens = rate
		}
		b.last = time.Now()
		if b.tokens < 1 {
			mu.Unlock()
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit", 429)
			return
		}
		b.tokens--
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
