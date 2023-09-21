package api

import (
	"net/http"
	"os"

	"github.com/stablecog/sc-go/auth/secure"
	"github.com/stablecog/sc-go/auth/store"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

func (a *ApiWrapper) UserAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	// Verify client id
	clientId := r.FormValue("client_id")
	_, err = store.GetCache().IsValidClientID(clientId)
	if err != nil {
		log.Infof("invalid client id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	redirectURI := r.FormValue("redirect_uri")
	if redirectURI == "" {
		log.Infof("redirect uri is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !utils.IsValidHTTPURL(redirectURI) {
		log.Infof("redirect uri is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	state := r.FormValue("state")
	if state == "" {
		log.Infof("state is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Infof("redirect uri %s", redirectURI)

	// Generate secure auth code
	code, err := secure.GenerateAuthCode(64)
	if err != nil {
		log.Errorf("Error generating auth code: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create request to store
	authReq := store.AuthorizationRequest{
		Code:        code,
		RedirectURI: redirectURI,
		State:       state,
	}

	// Save auth request in cache
	err = a.RedisStore.SaveAuthRequestInCache(&authReq)
	if err != nil {
		log.Errorf("Error saving auth request in cache: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// add query params to redirect uri
	redirectLocation, err := utils.AddQueryParam(os.Getenv("OAUTH_REDIRECT_BASE"), utils.QueryParam{Key: "app_code", Value: code}, utils.QueryParam{Key: "app_id", Value: clientId})
	if err != nil {
		log.Errorf("Error adding query params to redirect uri: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", redirectLocation)
	w.WriteHeader(http.StatusFound)
	return
}
