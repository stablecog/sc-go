package middleware

import (
	"net/http"
	"strings"

	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/utils"
)

func (m *Middleware) GeoIPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr, _ := r.Context().Value("user_id").(string)
			email, _ := r.Context().Value("user_email").(string)
			if utils.GetCountryCode(r) == "NZ" && !strings.HasSuffix(email, "@gmail.com") {
				// Webhook
				discord.FireGeoIPWebhook(utils.GetIPAddress(r), email, userIDStr)
			}

			next.ServeHTTP(w, r)
		})
	}
}
