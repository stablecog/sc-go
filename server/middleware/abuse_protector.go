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
	return time.Since(createdAt) < 24*time.Hour*14
}

var shouldBanRules []ShouldBanRule = []ShouldBanRule{
	{
		Reason: `From BD, has "+" or multiple dots in the email, new, and free.`,
		Func: func(r *http.Request) bool {
			email, _ := r.Context().Value("user_email").(string)
			activeProductID, _ := r.Context().Value("user_active_product_id").(string)
			createdAtStr, _ := r.Context().Value("user_created_at").(string)
			countryCode := utils.GetCountryCode(r)
			emailWithoutDomain := strings.Split(email, "@")[0]

			fromBD := countryCode == "BD"
			hasPlus := strings.Contains(emailWithoutDomain, "+")
			hasMultipleDots := strings.Count(emailWithoutDomain, ".") >= 2
			isFreeUser := activeProductID == ""
			isNew := isAccountNew(createdAtStr)

			shouldBan := fromBD && (hasPlus || hasMultipleDots) && isFreeUser && isNew
			return shouldBan
		},
	},
	{
		Reason: "Three dots in the address, @googlemail.com, new, and free.",
		Func: func(r *http.Request) bool {
			email, _ := r.Context().Value("user_email").(string)
			activeProductID, _ := r.Context().Value("user_active_product_id").(string)
			createdAtStr, _ := r.Context().Value("user_created_at").(string)
			emailWithoutDomain := strings.Split(email, "@")[0]

			hasThreeDots := strings.Count(emailWithoutDomain, ".") >= 3
			isGoogleMail := strings.HasSuffix(email, "@googlemail.com")
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
			emailWithoutDomain := strings.Split(email, "@")[0]

			isOutlook := strings.HasSuffix(email, "@outlook.com")
			hasPlus := strings.Contains(emailWithoutDomain, "+")
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
	{
		Reason: "Four dots in the address and Gmail.",
		Func: func(r *http.Request) bool {
			email, _ := r.Context().Value("user_email").(string)
			activeProductID, _ := r.Context().Value("user_active_product_id").(string)
			createdAtStr, _ := r.Context().Value("user_created_at").(string)
			emailWithoutDomain := strings.Split(email, "@")[0]

			hasFourDots := strings.Count(emailWithoutDomain, ".") >= 4
			isGoogleMail := strings.HasSuffix(email, "@googlemail.com")
			isGmail := strings.HasSuffix(email, "@gmail.com")
			isFreeUser := activeProductID == ""
			isNew := isAccountNew(createdAtStr)

			shouldBan := hasFourDots && (isGoogleMail || isGmail) && isFreeUser && isNew
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
				log.Warnf("AbuseProtectorMiddleware | Invalid email encountered while banning user: %s", email)
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

			if len(banReasons) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			user, err := m.Repo.GetUser(userID)

			if err != nil {
				log.Errorf("AbuseProtectorMiddleware | Error getting user: %s", err.Error())
				next.ServeHTTP(w, r)
				return
			}

			if user.BannedAt != nil {
				log.Infof(`AbuseProtectorMiddleware | User "%s" is already banned`, userIDStr)
				next.ServeHTTP(w, r)
				return
			}

			err = discord.FireBannedUserWebhook(utils.GetIPAddress(r), email, domain, userIDStr, utils.GetCountryCode(r), thumbmarkID, banReasons)
			if err != nil {
				log.Errorf("AbuseProtectorMiddleware | Error firing BannedUser webhook: %s", err.Error())
				next.ServeHTTP(w, r)
				return
			}

			// Ban the user
			_, err = m.Repo.BanUsers([]uuid.UUID{userID}, false)
			if err != nil {
				log.Errorf("AbuseProtectorMiddleware | Error updating user as banned: %s", err.Error())
			}
			time.Sleep(30 * time.Second)
			next.ServeHTTP(w, r)
			return
		})
	}
}
