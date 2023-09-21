package api

import (
	"context"
	"math"
	"net/http"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stablecog/sc-go/auth/store"
	"github.com/stablecog/sc-go/log"
)

func ClientFormHandler(r *http.Request) (string, string, error) {
	clientID := r.Form.Get("client_id")
	if clientID == "" {
		return "", "", errors.ErrInvalidClient
	}
	return clientID, "", nil
}

// ClientBasic

// GetAccessToken access token
func (a *ApiWrapper) GetAccessToken(ctx context.Context, s *server.Server, gt oauth2.GrantType, tgr *oauth2.TokenGenerateRequest) (oauth2.TokenInfo,
	error) {
	if allowed := s.CheckGrantType(gt); !allowed {
		return nil, errors.ErrUnauthorizedClient
	}

	// Verify client ID
	client, err := store.GetCache().IsValidClientID(tgr.ClientID)
	if err != nil {
		return nil, errors.ErrUnauthorizedClient
	}

	switch gt {
	case oauth2.AuthorizationCode:
		// Check store for valid code
		authApproval, err := a.RedisStore.GetAuthApproval(tgr.Code)
		if err != nil && err != redis.Nil {
			log.Errorf("Error getting auth approval from redis %v", err)
			return nil, errors.ErrServerError
		} else if err == redis.Nil {
			log.Errorf("Auth code not found in redis %s", tgr.Code)
			return nil, errors.ErrInvalidAuthorizeCode
		}

		// Decrypt
		decryptedAuthApproval, err := a.AesCrypt.Decrypt(authApproval)
		if err != nil {
			log.Errorf("Error decrypting auth approval %v", err)
			return nil, errors.ErrInvalidAuthorizeCode
		}

		userId, err := uuid.Parse(decryptedAuthApproval)
		if err != nil {
			log.Errorf("Error parsing auth approval uuid %v", err)
			return nil, errors.ErrInvalidAuthorizeCode
		}

		// Create token
		dbToken, tokenRaw, err := a.Repo.NewAPITokenForAuthClient(userId, client)
		ti := &models.Token{
			ClientID:        tgr.ClientID,
			UserID:          userId.String(),
			Scope:           "api",
			Code:            tgr.Code,
			Access:          tokenRaw,
			AccessCreateAt:  dbToken.CreatedAt,
			AccessExpiresIn: math.MaxInt64,
		}
		return ti, nil
	}

	return nil, errors.ErrUnsupportedGrantType
}
