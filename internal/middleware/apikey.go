package middleware

import (
	"net/http"

	"rmp-api/internal/config"
	"rmp-api/pkg/response"
)

func APIKey(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				response.Error(w, http.StatusUnauthorized, "missing API key")
				return
			}

			if key != cfg.APIKey {
				response.Error(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
