package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/discord"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

type RuleFunc func(*http.Request) bool

type ShouldBanRule struct {
	Reason string
	Func   RuleFunc
}

func isAccountNew(createdAtStr string) bool {
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		log.Errorf("Error parsing user created at: %s", err.Error())
		return false
	}
	return time.Since(createdAt) < 24*time.Hour*7
}

var shouldBanRules []ShouldBanRule = []ShouldBanRule{
	{
		Reason: "Three dots in the address, Gmail, new, and free.",
		Func: func(r *http.Request) bool {
			email, _ := r.Context().Value("user_email").(string)
			activeProductID, _ := r.Context().Value("user_active_product_id").(string)
			createdAtStr, _ := r.Context().Value("user_created_at").(string)

			hasThreeDots := strings.Count(email, ".") >= 4
			isGoogleMail := strings.HasSuffix(email, "@googlemail.com") || strings.HasSuffix(email, "@gmail.com")
			isFreeUser := activeProductID == ""
			isNew := isAccountNew(createdAtStr)

			shouldBan := hasThreeDots && isGoogleMail && isFreeUser && isNew
			return shouldBan
		},
	},
	{
		Reason: "Banned Thumbmark ID, new, and free.",
		Func: func(r *http.Request) bool {
			bannedThumbmarkIDs := shared.GetCache().ThumbmarkIDBlacklist()
			thumbmarkID, _ := r.Context().Value("user_thumbmark_id").(string)
			activeProductID, _ := r.Context().Value("user_active_product_id").(string)
			createdAtStr, _ := r.Context().Value("user_created_at").(string)

			isNew := isAccountNew(createdAtStr)
			isBannedThumbmarkID := false
			if thumbmarkID != "" && slices.Contains(bannedThumbmarkIDs, thumbmarkID) {
				isBannedThumbmarkID = true
			}
			isFreeUser := activeProductID == ""

			shouldBan := isBannedThumbmarkID && isFreeUser && isNew
			return shouldBan
		},
	},
	{
		Reason: "Has plus, three numbers, Outlook, new, and free.",
		Func: func(r *http.Request) bool {
			activeProductID, _ := r.Context().Value("user_active_product_id").(string)
			createdAtStr, _ := r.Context().Value("user_created_at").(string)
			email, _ := r.Context().Value("user_email").(string)

			isOutlook := strings.HasSuffix(email, "@outlook.com")
			hasPlus := strings.Contains(email, "+")
			isNew := isAccountNew(createdAtStr)
			isFreeUser := activeProductID == ""
			hasThreeNumbers := false

			regexPattern := `(?s)(.*\d.*){3,}`
			reg, err := regexp.Compile(regexPattern)
			if err != nil {
				log.Warnf("There was an error compiling the regex: %s", err.Error())
			} else {
				hasThreeNumbers = reg.MatchString(email)
			}

			shouldBan := isOutlook && hasPlus && hasThreeNumbers && isFreeUser && isNew
			return shouldBan
		},
	},
}

func (m *Middleware) AbuseProtectorMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr, _ := r.Context().Value("user_id").(string)
			userID, err := uuid.Parse(userIDStr)
			userIP := r.Context().Value("user_ip").(string)
			if err != nil {
				log.Errorf("Error parsing user ID: %s", err.Error())
				next.ServeHTTP(w, r)
				return
			}
			email, _ := r.Context().Value("user_email").(string)
			thumbmarkID, _ := r.Context().Value("user_thumbmark_id").(string)
			userActiveProductID, _ := r.Context().Value("user_active_product_id").(string)
			userBannedAt := r.Context().Value("user_banned_at").(string)
			isUserAlreadyBanned := userBannedAt != ""

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
				err = discord.FireBannedUserWebhook(utils.GetIPAddress(r), email, domain, userIDStr, utils.GetCountryCode(r), thumbmarkID, banReasons)
				if err != nil {
					log.Errorf("Error firing BannedUser webhook: %s", err.Error())
					next.ServeHTTP(w, r)
					return
				}
				// Ban the user if not banned already
				if !isUserAlreadyBanned {
					_, err = m.Repo.BanUsers([]uuid.UUID{userID}, false)
					if err != nil {
						log.Errorf("Error updating user as banned: %s", err.Error())
					} else {
						go m.Track.AutoBannedByAbuseProtector(userIDStr, email, userActiveProductID, userIP, banReasons)
					}
				}
				time.Sleep(30 * time.Second)
			}

			next.ServeHTTP(w, r)
		})
	}
}
