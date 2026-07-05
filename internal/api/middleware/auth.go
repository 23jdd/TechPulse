package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"techpulse/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func MockAuth(defaultUserID int64) func(http.Handler) http.Handler {
	return Auth(defaultUserID, "", false)
}

func Auth(defaultUserID int64, jwtSecret string, required bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := defaultUserID
			if token := bearerToken(r); token != "" {
				claims, err := auth.VerifyJWT(jwtSecret, token)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				userID = claims.Subject
			} else if required {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if raw := r.Header.Get("X-User-ID"); raw != "" && !required {
				if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
					userID = parsed
				}
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDKey, userID)))
		})
	}
}

func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func UserID(r *http.Request) int64 {
	if value, ok := r.Context().Value(UserIDKey).(int64); ok && value > 0 {
		return value
	}
	return 1
}
