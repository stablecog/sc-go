package rest

import (
	"net/http"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/clip"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/server/scworker"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	stripe "github.com/stripe/stripe-go/v74/client"
)

// Shared pagination defaults
const DEFAULT_PER_PAGE = 50
const MAX_PER_PAGE = 300

type RestAPI struct {
	Repo           *repository.Repository
	Redis          *database.RedisWrapper
	Hub            *sse.Hub
	StripeClient   *stripe.API
	Track          *analytics.AnalyticsService
	QueueThrottler *shared.UserQueueThrottlerMap
	S3             *s3.S3
	Qdrant         *qdrant.QdrantClient
	Clip           *clip.ClipService
	SafetyChecker  *utils.TranslatorSafetyChecker
	SCWorker       *scworker.SCWorker
	// For API key requests to track them
	SMap *shared.SyncMap[chan requests.CogWebhookMessage]
}

func (c *RestAPI) GetApiToken(w http.ResponseWriter, r *http.Request) (token *ent.ApiToken) {
	// Get token from ctx
	tokenIdStr, ok := r.Context().Value("api_token_id").(string)
	if !ok || tokenIdStr == "" {
		responses.ErrUnauthorized(w, r)
		return nil
	}
	// Ensure valid uuid
	tokenId, err := uuid.Parse(tokenIdStr)
	if err != nil {
		responses.ErrUnauthorized(w, r)
		return nil
	}

	// get from DB
	t, err := c.Repo.GetToken(tokenId)
	if err != nil || t == nil {
		log.Error("Error getting token", "err", err)
		responses.ErrUnauthorized(w, r)
		return nil
	}
	return t
}

func (c *RestAPI) GetUserIfAuthenticated(w http.ResponseWriter, r *http.Request) (user *ent.User) {
	// See if authenticated
	userIDStr, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userIDStr == "" {
		responses.ErrUnauthorized(w, r)
		return nil
	}
	// Ensure valid uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		responses.ErrUnauthorized(w, r)
		return nil
	}

	// Get user
	user, err = c.Repo.GetUser(userID)
	if err != nil || user == nil {
		log.Error("Error getting user", "err", err)
		responses.ErrUnauthorized(w, r)
		return nil
	}
	return user
}

func (c *RestAPI) GetUserIDAndEmailIfAuthenticated(w http.ResponseWriter, r *http.Request) (id *uuid.UUID, email string) {
	// See if authenticated
	userIDStr, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userIDStr == "" {
		responses.ErrUnauthorized(w, r)
		return nil, ""
	}
	// Ensure valid uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		responses.ErrUnauthorized(w, r)
		return nil, ""
	}

	// Get email
	email, ok := r.Context().Value("user_email").(string)
	if !ok {
		responses.ErrUnauthorized(w, r)
		return nil, ""
	}
	return &userID, email
}
