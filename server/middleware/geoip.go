package middleware

import (
	"net/http"

	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/utils"
)

func (m *Middleware) GeoIPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr, _ := r.Context().Value("user_id").(string)
			email, _ := r.Context().Value("user_email").(string)
			country, err := m.GeoIP.GetCountryFromIP(utils.GetIPAddress(r))
			if err != nil {
				log.Warn("Error getting country from IP", "err", err)
			} else {
				if country == "TR" {
					// Webhook
					discord.FireGeoIPWebhook(utils.GetIPAddress(r), email, userIDStr)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
