package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/server/responses"
	stripe "github.com/stripe/stripe-go/client"
)

type RestAPI struct {
	Repo         *repository.Repository
	Redis        *database.RedisWrapper
	Hub          *sse.Hub
	StripeClient *stripe.API
	Meili        *meilisearch.Client
}

func (c *RestAPI) GetUserIDIfAuthenticated(w http.ResponseWriter, r *http.Request) *uuid.UUID {
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
	return &userID
}
