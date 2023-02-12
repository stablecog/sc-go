package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/server/responses"
)

// Enforces authorization
func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			responses.ErrUnauthorized(w, r)
			return
		}
		// Check supabase to see if it's all good
		start := time.Now()
		userId, err := m.SupabaseAuth.GetSupabaseUserIdFromAccessToken(authHeader[1])
		if err != nil {
			responses.ErrUnauthorized(w, r)
			return
		}
		fmt.Printf("--- Get supabase user_id took: %s\n", time.Now().Sub(start))
		// Set the user ID in the context
		ctx := context.WithValue(r.Context(), "user_id", userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Enforces super_admin or gallery_admin
// MUST be called after AuthMiddleware
func (m *Middleware) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIdStr, ok := r.Context().Value("user_id").(string)
		if !ok {
			responses.ErrUnauthorized(w, r)
			return
		}
		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			responses.ErrUnauthorized(w, r)
			return
		}

		roles, err := m.Repo.GetRoles(userID)
		if err != nil {
			responses.ErrUnauthorized(w, r)
			return
		}

		for _, role := range roles {
			if role == userrole.RoleNameSUPER_ADMIN || role == userrole.RoleNameGALLERY_ADMIN {
				next.ServeHTTP(w, r)
				return
			}
		}

		responses.ErrUnauthorized(w, r)
		return
	})
}

// Enforces super_admin
// MUST be called after AuthMiddleware
func (m *Middleware) SuperAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIdStr, ok := r.Context().Value("user_id").(string)
		if !ok {
			responses.ErrUnauthorized(w, r)
			return
		}
		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			responses.ErrUnauthorized(w, r)
			return
		}

		superAdmin, err := m.Repo.IsSuperAdmin(userID)
		if err != nil || !superAdmin {
			responses.ErrUnauthorized(w, r)
			return
		}

		next.ServeHTTP(w, r)
		return
	})
}
