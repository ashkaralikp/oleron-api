package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"rmp-api/pkg/response"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

// ClientIP resolves the originating client IP for logging, rate limiting, and persistence.
func ClientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

// RateLimit allows maxRequests per window per IP
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	visitors := make(map[string]*visitor)
	var mu sync.Mutex

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
			ip := ClientIP(r)
			if ip == "" {
				ip = "unknown"
			}

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
