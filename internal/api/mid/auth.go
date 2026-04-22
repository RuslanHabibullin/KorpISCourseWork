package mid

import (
	"net/http"
	"strings"
)

// Auth — простой middleware проверки Bearer-токена.
// В production заменить на JWT-валидацию.
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				// Auth отключён
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] != token {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
