package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Whitelist email domains
var whitelist = []string{
	"gmail.com",
	"yahoo.com",
	"yahoo.de",
	"yahoo.fr",
	"yahoo.co.uk",
	"yahoo.it",
	"hotmail.com",
	"outlook.de",
	"outlook.com",
	"icloud.com",
	"googlemail.com",
	"proton.me",
	"protonmail.com",
	"qq.com",
	"gmx.de",
	"mail.ru",
	"yandex.ru",
	"live.com",
	"live.ca",
	"aol.com",
	"hotmail.co.uk",
	"hotmail.fr",
	"mail.com",
	"me.com",
	"gmx.net",
	"hotmail.de",
	"wp.pl",
	"bluewin.ch",
	"icloud.com",
	"libero.it",
	"bk.ru",
	"comcast.net",
}

func (m *Middleware) GeoIPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr, _ := r.Context().Value("user_id").(string)
			email, _ := r.Context().Value("user_email").(string)
			thumbmarkID, _ := r.Context().Value("thumbmark_id").(string)
			if utils.GetCountryCode(r) == "BD" {
				// See if not in whitelist
				whitelisted := false
				for _, domain := range whitelist {
					if strings.HasSuffix(email, "@"+domain) {
						whitelisted = true
						break
					}
				}
				if !whitelisted {
					// Get domain
					segs := strings.Split(email, "@")
					if len(segs) != 2 {
						log.Warnf("Invalid email encountered in GeoIP: %s", email)
						next.ServeHTTP(w, r)
						return
					}
					domain := strings.ToLower(segs[1])
					// Webhook
					err := discord.FireGeoIPBannedUserWebhook(utils.GetIPAddress(r), email, domain, userIDStr, utils.GetCountryCode(r), thumbmarkID)
					if err != nil {
						log.Errorf("Error firing GeoIP webhook: %s", err.Error())
						next.ServeHTTP(w, r)
						return
					}
					// Insert into disposable email domains
					_, err = m.Repo.BanDomains([]string{domain}, false)
					if err != nil {
						log.Errorf("Error inserting disposable email domain: %s", err.Error())
					} else {
						// Update in cache immediately
						shared.GetCache().UpdateDisposableEmailDomains(append(shared.GetCache().DisposableEmailDomains(), domain))
					}
					// Sleep 30 seconds
					time.Sleep(30 * time.Second)
				}
			}

			if utils.GetCountryCode(r) == "TR" {
				segs := strings.Split(email, "@")
				if len(segs) != 2 {
					log.Warnf("Invalid email encountered in GeoIP: %s", email)
					next.ServeHTTP(w, r)
					return
				}
				domain := strings.ToLower(segs[1])
				err := discord.FireGeoIPSuspiciousUserWebhook(utils.GetIPAddress(r), email, domain, userIDStr, utils.GetCountryCode(r), thumbmarkID)
				if err != nil {
					log.Errorf("Error firing GeoIP webhook: %s", err.Error())
					next.ServeHTTP(w, r)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
