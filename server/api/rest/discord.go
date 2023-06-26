package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
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
	token, _ := c.Redis.GetDiscordTokenFromID(authReq.DiscordID)
	if token != authReq.Token {
		responses.ErrUnauthorized(w, r)
		return
	}

	// See if user already exists with this ID
	_, err = c.Repo.GetUserByDiscordID(authReq.DiscordID)
	if err != nil && ent.IsNotFound(err) {
		responses.ErrBadRequest(w, r, "already_linked", "This discord account is already linked to a Stablecog account")
		return
	} else if err != nil {
		log.Errorf("Error checking for existing discord user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Update user with discord ID
	err = c.Repo.SetDiscordID(user.ID, authReq.DiscordID)
	if err != nil {
		log.Errorf("Error setting discord ID: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Delete token from redis
	err = c.Redis.DeleteDiscordToken(authReq.DiscordID)
	if err != nil {
		log.Warnf("Error deleting discord token: %v", err)
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]bool{"success": true})
}
