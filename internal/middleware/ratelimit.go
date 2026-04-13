package middleware

import (
	"net/http"
	"sync"
	"time"

	"rmp-api/pkg/response"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

// RateLimit allows maxRequests per window per IP
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	// Cleanup stale entries periodically
	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			v, exists := visitors[ip]
			if !exists || time.Since(v.lastSeen) > window {
				visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			v.count++
			v.lastSeen = time.Now()

			if v.count > maxRequests {
				mu.Unlock()
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
