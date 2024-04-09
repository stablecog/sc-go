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

type ShouldBanRule struct {
	Reason string
	Func   RuleFunc
}

var shouldBanRules []ShouldBanRule = []ShouldBanRule{
	{
		Reason: "Free, Gmail, and 3 dots in the address.",
		Func: func(r *http.Request) bool {
			userIsPaying := r.Context().Value("user_is_paying").(bool)
			email, _ := r.Context().Value("user_email").(string)
			hasThreeDots := strings.Count(email, ".") >= 4
			isGoogleMail := strings.HasSuffix(email, "@googlemail.com") || strings.HasSuffix(email, "@gmail.com")
			shouldBan := hasThreeDots && isGoogleMail && !userIsPaying
			return shouldBan
		},
	},
}

func (m *Middleware) AbuseProtectorMiddleware() func(next http.Handler) http.Handler {
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
				log.Warnf("Invalid email encountered while banning user: %s", email)
				next.ServeHTTP(w, r)
				return
			}
			domain := strings.ToLower(segs[1])

			banReasons := []string{}
			for _, shouldBan := range shouldBanRules {
				if shouldBan.Func(r) {
					banReasons = append(banReasons, shouldBan.Reason)
				}
			}

			if len(banReasons) > 0 {
				// Webhook
				err := discord.FireBannedUserWebhook(utils.GetIPAddress(r), email, domain, userIDStr, utils.GetCountryCode(r), thumbmarkID, banReasons)
				if err != nil {
					log.Errorf("Error firing BannedUser webhook: %s", err.Error())
					next.ServeHTTP(w, r)
					return
				}
				// Ban the user
				_, err = m.Repo.BanUsers([]uuid.UUID{userID}, false)
				if err != nil {
					log.Errorf("Error updating user as banned: %s", err.Error())
				}
				time.Sleep(30 * time.Second)
			}

			next.ServeHTTP(w, r)
		})
	}
}
