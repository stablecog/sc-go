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

type AuthLevel int

const (
	AuthLevelAny AuthLevel = iota
	AuthLevelGalleryAdmin
	AuthLevelSuperAdmin
)

// Enforces authorization at specific level
func (m *Middleware) AuthMiddleware(level AuthLevel) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			userIDParsed, err := uuid.Parse(userId)
			if err != nil {
				responses.ErrUnauthorized(w, r)
				return
			}

			// They do have an appropriate auth token
			authorized := true

			// If level is more than any, check if they have the appropriate role
			if level == AuthLevelGalleryAdmin || level == AuthLevelSuperAdmin {
				authorized = false
				roles, err := m.Repo.GetRoles(userIDParsed)
				if err != nil {
					responses.ErrUnauthorized(w, r)
					return
				}
				for _, role := range roles {
					// Super admin always authorized
					if role == userrole.RoleNameSUPER_ADMIN {
						authorized = true
						break
					} else if role == userrole.RoleNameGALLERY_ADMIN && level == AuthLevelGalleryAdmin {
						// Gallery admin only authorized if we're checking for gallery admin
						authorized = true
						break
					}
				}
			}

			if !authorized {
				responses.ErrUnauthorized(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
