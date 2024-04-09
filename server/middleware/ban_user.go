package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/utils"
)

type RuleFunc func(*http.Request) bool

var shouldBanRules []RuleFunc = []RuleFunc{
	func(r *http.Request) bool {
		email, _ := r.Context().Value("user_email").(string)
		hasThreeDots := strings.Count(email, ".") >= 4
		isGoogleMail := strings.HasSuffix(email, "@googlemail.com")
		shouldBan := hasThreeDots && isGoogleMail
		return shouldBan
	},
}

func (m *Middleware) BanUserMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr, _ := r.Context().Value("user_id").(string)
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				log.Errorf("Error parsing user ID: %s", err.Error())
				next.ServeHTTP(w, r)
				return
			}
			email, _ := r.Context().Value("user_email").(string)
			thumbmarkID, _ := r.Context().Value("user_thumbmark_id").(string)

			// Get domain
			segs := strings.Split(email, "@")
			if len(segs) != 2 {
				log.Warnf("Invalid email encountered in GeoIP: %s", email)
				next.ServeHTTP(w, r)
				return
			}
			domain := strings.ToLower(segs[1])

			banReasonCounter := 0
			for _, shouldBan := range shouldBanRules {
				if shouldBan(r) {
					banReasonCounter++
				}
			}

			if banReasonCounter > 0 {
				// Webhook
				err := discord.FireBannedUserWebhook(utils.GetIPAddress(r), email, domain, userIDStr, utils.GetCountryCode(r), thumbmarkID)
				if err != nil {
					log.Errorf("Error firing BannedUser webhook: %s", err.Error())
					next.ServeHTTP(w, r)
					return
				}
				// Ban the user
				_, err = m.Repo.BanUsers([]uuid.UUID{userID}, false)
				if err != nil {
					log.Errorf("Error inserting user into banned users: %s", err.Error())
				}
				time.Sleep(30 * time.Second)
			}

			next.ServeHTTP(w, r)
		})
	}
}
