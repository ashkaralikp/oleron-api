package middleware

import (
	"net/http"

	"clinic-api/pkg/response"
)

// RequireRole restricts access to users with one of the allowed roles.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok || role == "" {
				response.Error(w, http.StatusForbidden, "access denied")
				return
			}

			for _, a := range allowed {
				if role == a {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.Error(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}
