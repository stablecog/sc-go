package middleware

import (
	"context"
	"crypto/subtle"
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
			userId, email, lastSignIn, err := m.SupabaseAuth.GetSupabaseUserIdFromAccessToken(authHeader[1])
			if err != nil {
				responses.ErrUnauthorized(w, r)
				return
			}

			// Set the user ID in the context
			ctx := context.WithValue(r.Context(), "user_id", userId)
			ctx = context.WithValue(ctx, "user_email", email)
			// Set the last sign in time in the context, if not null
			if lastSignIn != nil {
				formatted := lastSignIn.Format(time.RFC3339)
				ctx = context.WithValue(ctx, "user_last_sign_in", formatted)
			}

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
						ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())
						break
					} else if role == userrole.RoleNameGALLERY_ADMIN && level == AuthLevelGalleryAdmin {
						// Gallery admin only authorized if we're checking for gallery admin
						authorized = true
						ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())
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

// BasicAuth is a wrapper for Handler that requires username and password
func BasicAuth(next http.Handler, username, password, realm string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
