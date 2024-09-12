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
	AuthLevelOptional
)

// Enforces authorization at specific level
func (m *Middleware) AuthMiddleware(levels ...AuthLevel) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Copy the levels array
			levelsCopy := make([]AuthLevel, len(levels))
			copy(levelsCopy, levels)
			authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(authHeader) != 2 {
				if slices.Contains(levelsCopy, AuthLevelOptional) {
					next.ServeHTTP(w, r)
					return
				}
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

			// Separate flow for API tokens
			if slices.Contains(levelsCopy, AuthLevelOptional) {
				if strings.HasPrefix(authHeader[1], "sc-") && len(authHeader[1]) == 67 {
					levelsCopy = append(levelsCopy, AuthLevelAPIToken)
				}
			}
			if slices.Contains(levelsCopy, AuthLevelGalleryAdmin) || slices.Contains(levelsCopy, AuthLevelSuperAdmin) {
				if strings.HasPrefix(authHeader[1], "sc-") && len(authHeader[1]) == 67 {
					levelsCopy = append(levelsCopy, AuthLevelAPIToken)
				}
			}
			if slices.Contains(levelsCopy, AuthLevelAPIToken) {
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

				if slices.Contains(levelsCopy, AuthLevelGalleryAdmin) || slices.Contains(levelsCopy, AuthLevelSuperAdmin) {
					authorized := false
					roles, err := m.Repo.GetRoles(user.ID)
					if err != nil {
						log.Error("Error getting roles", "err", err)
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
				userId, email, lastSignIn, err = m.SupabaseAuth.GetSupabaseUserIdFromAccessToken(authHeader[1], m.Repo.DB, m.Repo.Ctx)
				if err != nil {
					log.Error("Error getting user id from access token", "err", err)
					responses.ErrUnauthorized(w, r)
					return
				}
			}

			userIDParsed, err := uuid.Parse(userId)
			if err != nil {
				// This should never happen
				log.Error("Error parsing user ID", "err", err)
				responses.ErrUnauthorized(w, r)
				return
			}

			thumbmarkID := utils.GetThumbmarkID(r)

			user, err := m.Repo.GetUser(userIDParsed)
			if err != nil {
				log.Error("Error getting user", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error has occured")
				return
			}

			activeProductIDStr := ""
			createdAtStr := ""

			if user != nil && user.ActiveProductID != nil {
				activeProductIDStr = *user.ActiveProductID
			}

			if user != nil {
				createdAtStr = user.CreatedAt.Format(time.RFC3339)
			}

			// Set the user ID in the context
			ctx = context.WithValue(ctx, "user_id", userId)
			ctx = context.WithValue(ctx, "user_email", email)
			ctx = context.WithValue(ctx, "user_thumbmark_id", thumbmarkID)
			ctx = context.WithValue(ctx, "user_active_product_id", activeProductIDStr)
			ctx = context.WithValue(ctx, "user_created_at", createdAtStr)
			ctx = context.WithValue(ctx, "user_ip", ip)

			// Set the last sign in time in the context, if not null
			if lastSignIn != nil {
				formatted := lastSignIn.Format(time.RFC3339)
				ctx = context.WithValue(ctx, "user_last_sign_in", formatted)
			}

			// They do have an appropriate auth token
			authorized := true

			// If level is more than any, check if they have the appropriate role
			if slices.Contains(levelsCopy, AuthLevelGalleryAdmin) || slices.Contains(levelsCopy, AuthLevelSuperAdmin) {
				authorized = false
				roles, err := m.Repo.GetRoles(userIDParsed)
				if err != nil {
					log.Error("Error getting roles - 2", "err", err)
					responses.ErrUnauthorized(w, r)
					return
				}
				for _, role := range roles {
					// Super admin always authorized
					if role == "SUPER_ADMIN" {
						authorized = true
						ctx = context.WithValue(ctx, "user_role", "SUPER_ADMIN")
						break
					} else if role == "GALLERY_ADMIN" && slices.Contains(levelsCopy, AuthLevelGalleryAdmin) {
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
