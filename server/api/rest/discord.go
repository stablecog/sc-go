package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
)

// POST Discord Verification
// Handles linking a discord account to a Stablecog account
func (c *RestAPI) HandleAuthorizeDiscord(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var authReq requests.DiscordVerificationRequest
	err := json.Unmarshal(reqBody, &authReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Verify token exists in redis and is valid
	token, err := c.Redis.GetDiscordTokenFromID(authReq.UserID)
	if err != nil {
		if err == redis.Nil {
			responses.ErrBadRequest(w, r, "invalid_token", "The token provided is invalid")
			return
		} else {
			log.Errorf("Error getting discord token from redis: %v", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}
	}
	if token != authReq.Token {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, &responses.ErrorResponse{
			Error: "invalid_token",
		})
		return
	}

	// See if user already exists with this ID
	u, err := c.Repo.GetUserByDiscordID(authReq.UserID)
	if err != nil && !ent.IsNotFound(err) {
		log.Errorf("Error checking for existing discord user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	if u != nil {
		responses.ErrBadRequest(w, r, "already_linked", "This discord account is already linked to a Stablecog account")
		return
	}

	// Update user with discord ID and collect unclaimed tips
	var claimedTips int32
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()
		err = c.Repo.SetDiscordID(user.ID, authReq.UserID, db)
		if err != nil {
			return err
		}

		claimedTips, err = c.Repo.CollectUnclaimedTips(user.ID, authReq.UserID, db)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Errorf("Errir in discord verification transaction: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Delete token from redis
	err = c.Redis.DeleteDiscordToken(authReq.UserID)
	if err != nil {
		log.Warnf("Error deleting discord token: %v", err)
	}

	// Publish connected message to disco bot
	discordAuthMsg := responses.DiscordRedisStreamMessage{
		DiscordId: authReq.UserID,
	}
	respBytes, err := json.Marshal(discordAuthMsg)
	if err == nil {
		c.Redis.Client.Publish(c.Redis.Ctx, shared.REDIS_DISCORD_COG_CHANNEL, respBytes)
	} else {
		log.Error("Error marshalling disco bot connection message", "err", err)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"success": true, "tips_claimed": claimedTips})
}
