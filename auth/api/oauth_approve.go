package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/redis/go-redis/v9"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

type ApproveAuthorizationRequest struct {
	// The authorization code
	Code string `json:"app_code"`
}

type ApproveAuthorizationResponse struct {
	RedirectURL string `json:"redirect_url"`
}

// When the user approves this authorization request from the UI
func (a *ApiWrapper) ApproveAuthorization(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var approveReq ApproveAuthorizationRequest
	err := json.Unmarshal(reqBody, &approveReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// See if code is valid
	authReq, err := a.RedisStore.GetAuthRequestFromCache(approveReq.Code)
	if err != nil && err != redis.Nil {
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	} else if err == redis.Nil {
		log.Error("Invalid auth code", "code", approveReq.Code)
		responses.ErrUnauthorized(w, r)
		return
	}

	// Get bearer token from header
	authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(authHeader) != 2 {
		responses.ErrUnauthorized(w, r)
		return
	}

	accessToken := authHeader[1]

	// Verify token
	id, _, _, err := a.SupabaseAuth.GetSupabaseUserIdFromAccessToken(accessToken)
	if err != nil {
		log.Errorf("Error logging into supabase %v", err)
		responses.ErrUnauthorized(w, r)
		return
	}

	// Store encrypted user_id
	encryptedUserId, err := a.AesCrypt.Encrypt(id)
	if err != nil {
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	err = a.RedisStore.StoreAuthApproval(authReq.Code, encryptedUserId)

	// Return redirect url
	redirectUrl, err := utils.AddQueryParam(authReq.RedirectURI, utils.QueryParam{Key: "code", Value: authReq.Code}, utils.QueryParam{Key: "state", Value: authReq.State})
	if err != nil {
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Clear auth request from cache
	a.RedisStore.ClearAuthRequestFromCache(approveReq.Code)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, ApproveAuthorizationResponse{
		RedirectURL: redirectUrl,
	})
}
