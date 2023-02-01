package controller

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/controller/websocket"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
)

type HttpController struct {
	Repo                       *repository.Repository
	Redis                      *database.RedisWrapper
	S3Client                   *s3.Client
	S3PresignClient            *s3.PresignClient
	CogRequestWebsocketConnMap *shared.SyncMap[string]
	Hub                        *websocket.Hub
}

func (c *HttpController) GetUserIDIfAuthenticated(w http.ResponseWriter, r *http.Request) *uuid.UUID {
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
