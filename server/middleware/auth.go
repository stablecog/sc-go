package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/stablecog/go-apps/server/models"
)

// Enforces authorization
func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			models.ErrUnauthorized(w, r)
			return
		}
		// Check supabase to see if it's all good
		userId, err := m.SupabaseAuth.GetSupabaseUserIdFromAccessToken(authHeader[1])
		if err != nil {
			models.ErrUnauthorized(w, r)
			return
		}
		// Set the user ID in the context
		ctx := context.WithValue(r.Context(), "user_id", userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Optional authorization (sets user_id in context if present)
func (m *Middleware) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			next.ServeHTTP(w, r)
			return
		}
		// Check supabase to see if it's all good
		userId, err := m.SupabaseAuth.GetSupabaseUserIdFromAccessToken(authHeader[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		// Set the user ID in the context
		ctx := context.WithValue(r.Context(), "user_id", userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
