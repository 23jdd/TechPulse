package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func MockAuth(defaultUserID int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := defaultUserID
			if raw := r.Header.Get("X-User-ID"); raw != "" {
				if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
					userID = parsed
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDKey, userID)))
		})
	}
}

func UserID(r *http.Request) int64 {
	if value, ok := r.Context().Value(UserIDKey).(int64); ok && value > 0 {
		return value
	}
	return 1
}
