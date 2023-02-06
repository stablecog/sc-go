package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/api/websocket"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
)

type RestAPI struct {
	Repo                       *repository.Repository
	Redis                      *database.RedisWrapper
	CogRequestWebsocketConnMap *shared.SyncMap[string]
	Hub                        *websocket.Hub
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
