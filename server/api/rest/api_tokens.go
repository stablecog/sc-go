package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
)

// GET - Get active API tokens for user
func (c *RestAPI) HandleGetAPITokens(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	tokens, err := c.Repo.GetTokensByUserID(user.ID, true)
	if err != nil {
		log.Error("Error getting tokens", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	tokenRes := make([]responses.ApiToken, len(tokens))
	for i, token := range tokens {
		tokenRes[i] = responses.ApiToken{
			ID:           token.ID,
			Name:         token.Name,
			ShortString:  token.ShortString,
			Uses:         token.Uses,
			CreditsSpent: token.CreditsSpent,
			IsActive:     token.IsActive,
			LastUsedAt:   token.LastUsedAt,
			CreatedAt:    token.CreatedAt,
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.GetApiTokensResponse{
		Tokens: tokenRes,
	})
}

// POST - Create a new API token for the user
func (c *RestAPI) HandleNewAPIToken(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var newReq requests.NewTokenRequest
	err := json.Unmarshal(reqBody, &newReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Truncate to max length
	if len(newReq.Name) > shared.MAX_TOKEN_NAME_SIZE {
		newReq.Name = newReq.Name[:shared.MAX_TOKEN_NAME_SIZE]
	}

	// See if user already has more than max tokens
	count, err := c.Repo.GetTokenCountByUserID(user.ID)
	if err != nil {
		log.Error("Error getting token count", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	if count >= shared.MAX_API_TOKENS_PER_USER {
		responses.ErrBadRequest(w, r, "too_many_tokens", fmt.Sprintf("You already have the maximum number of API tokens (%d)", shared.MAX_API_TOKENS_PER_USER))
		return
	}

	// Create new token
	token, tokenStr, err := c.Repo.NewAPIToken(user.ID, newReq.Name)
	if err != nil {
		log.Error("Error creating new token", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	res := responses.NewApiTokensResponse{
		ID:    token.ID,
		Token: tokenStr,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// POST - Create a new API token for the user
func (c *RestAPI) HandleDeactivateAPIToken(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var deactivateReq requests.DeactiveApiTokenRequest
	err := json.Unmarshal(reqBody, &deactivateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Deactivate
	count, err := c.Repo.DeactivateTokenForUser(deactivateReq.ID, user.ID)
	if err != nil {
		log.Error("Error deactivating token", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	if count == 0 {
		responses.ErrNotFound(w, r, "token_not_found")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
}
