package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

type AuthLevel int

const (
	AuthLevelAny AuthLevel = iota
	AuthLevelGalleryAdmin
	AuthLevelSuperAdmin
	AuthLevelAPIToken
)

// Enforces authorization at specific level
func (m *Middleware) AuthMiddleware(levels ...AuthLevel) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(authHeader) != 2 {
				responses.ErrUnauthorized(w, r)
				return
			}

			ip := utils.GetIPAddress(r)
			if shared.GetCache().IsIPBanned(ip) {
				responses.ErrUnauthorized(w, r)
				return
			}

			var userId, email string
			var lastSignIn *time.Time
			var err error
			ctx := r.Context()

			// Separe flow for API tokens
			if slices.Contains(levels, AuthLevelAPIToken) {
				// Hash token
				hashed := utils.Sha256(authHeader[1])
				// Validate
				token, err := m.Repo.GetTokenByHashedToken(hashed)
				if err != nil && !ent.IsNotFound(err) {
					log.Error("Error getting token", "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error has occured")
					return
				} else if ent.IsNotFound(err) || !token.IsActive {
					responses.ErrUnauthorized(w, r)
					return
				}

				user, err := m.Repo.GetUser(token.UserID)
				if err != nil {
					log.Error("Error getting user", "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error has occured")
					return
				}

				if slices.Contains(levels, AuthLevelGalleryAdmin) || slices.Contains(levels, AuthLevelSuperAdmin) {
					authorized := false
					roles, err := m.Repo.GetRoles(user.ID)
					if err != nil {
						responses.ErrUnauthorized(w, r)
						return
					}
					for _, role := range roles {
						// Super admin always authorized
						if role == "SUPER_ADMIN" {
							authorized = true
							ctx = context.WithValue(ctx, "user_role", "SUPER_ADMIN")
							break
						} else if role == "GALLERY_ADMIN" && slices.Contains(levels, AuthLevelGalleryAdmin) {
							// Gallery admin only authorized if we're checking for gallery admin
							authorized = true
							ctx = context.WithValue(ctx, "user_role", "GALLERY_ADMIN")
						}
					}

					if !authorized {
						responses.ErrUnauthorized(w, r)
						return
					}
				}

				userId = user.ID.String()
				email = user.Email
				lastSignIn = user.LastSignInAt
				ctx = context.WithValue(ctx, "api_token_id", token.ID.String())
			} else {
				// Check supabase to see if it's all good
				userId, email, lastSignIn, err = m.SupabaseAuth.GetSupabaseUserIdFromAccessToken(authHeader[1])
				if err != nil {
					responses.ErrUnauthorized(w, r)
					return
				}
			}

			// Set the user ID in the context
			ctx = context.WithValue(ctx, "user_id", userId)
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
			if slices.Contains(levels, AuthLevelGalleryAdmin) || slices.Contains(levels, AuthLevelSuperAdmin) {
				authorized = false
				roles, err := m.Repo.GetRoles(userIDParsed)
				if err != nil {
					responses.ErrUnauthorized(w, r)
					return
				}
				for _, role := range roles {
					// Super admin always authorized
					if role == "SUPER_ADMIN" {
						authorized = true
						ctx = context.WithValue(ctx, "user_role", "SUPER_ADMIN")
						break
					} else if role == "GALLERY_ADMIN" && slices.Contains(levels, AuthLevelGalleryAdmin) {
						// Gallery admin only authorized if we're checking for gallery admin
						authorized = true
						ctx = context.WithValue(ctx, "user_role", "GALLERY_ADMIN")
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
