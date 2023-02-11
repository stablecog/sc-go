package repository

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/userrole"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

func (r *Repository) IsSuperAdmin(userID uuid.UUID) (bool, error) {
	// Check for admin
	roles, err := r.GetRoles(userID)
	if err != nil {
		klog.Errorf("Error getting user roles: %v", err)
		return false, err
	}
	for _, role := range roles {
		if role == userrole.RoleNameSUPER_ADMIN {
			return true, nil
		}
	}

	return false, nil
}

func (r *Repository) GetRoles(userID uuid.UUID) ([]userrole.RoleName, error) {
	roles, err := r.DB.UserRole.Query().Where(userrole.UserIDEQ(userID)).All(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting user roles: %v", err)
		return nil, err
	}
	var roleNames []userrole.RoleName
	for _, role := range roles {
		roleNames = append(roleNames, role.RoleName)
	}

	return roleNames, nil
}

// Process a cog message into database
func (r *Repository) ProcessCogMessage(msg responses.CogStatusUpdate) {
	redisKey := fmt.Sprintf("first:%s", msg.Input.ID)
	if msg.Status != responses.CogProcessing {
		redisKey = fmt.Sprintf("second:%s", msg.Input.ID)
	}
	// Since we sync with other instances, we get the stream ID from redis
	streamIdStr, err := r.Redis.GetCogRequestStreamID(r.Redis.Client.Context(), redisKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Probably means another instance picked this up
			return
		}
		klog.Errorf("--- Error getting stream ID from redis: %v", err)
		return
	}

	// We delete this, if our delete is successful then that means we are the ones responsible for processing it
	deleted, err := r.Redis.DeleteCogRequestStreamID(r.Redis.Client.Context(), redisKey)
	if err != nil {
		klog.Errorf("--- Error deleting stream ID from redis: %v", err)
		return
	}
	if deleted == 0 {
		// Means another instance already deleted it probably and handled it
		return
	}

	// Ensure is valid
	if !utils.IsSha256Hash(streamIdStr) {
		// Not sure how we get here, we never should
		klog.Errorf("--- Invalid SSE stream id: %s", streamIdStr)
		return
	}

	// Get process type
	if msg.Input.ProcessType != shared.GENERATE && msg.Input.ProcessType != shared.UPSCALE && msg.Input.ProcessType != shared.GENERATE_AND_UPSCALE {
		klog.Errorf("--- Invalid process type from cog, can't handle message: %s", msg.Input.ProcessType)
		return
	}

	var upscaleOutput *ent.UpscaleOutput
	var generationOutputs []*ent.GenerationOutput
	var cogErr string

	// Handle started/failed/succeeded message types
	if msg.Status == responses.CogProcessing {
		// In goroutine since we want them to know it started asap
		if msg.Input.ProcessType == shared.UPSCALE {
			go r.SetUpscaleStarted(msg.Input.ID)
		} else {
			go r.SetGenerationStarted(msg.Input.ID)
		}
	} else if msg.Status == responses.CogFailed {
		if msg.Input.ProcessType == shared.UPSCALE {
			r.SetUpscaleFailed(msg.Input.ID, msg.Error)
		} else {
			r.SetGenerationFailed(msg.Input.ID, msg.Error)
		}
		cogErr = msg.Error
	} else if msg.Status == responses.CogSucceeded {
		if len(msg.Outputs) == 0 {
			cogErr = "No outputs"
			if msg.Input.ProcessType == shared.UPSCALE {
				r.SetUpscaleFailed(msg.Input.ID, cogErr)
			} else {
				r.SetGenerationFailed(msg.Input.ID, cogErr)
			}
			msg.Status = responses.CogFailed
		} else {
			if msg.Input.ProcessType == shared.UPSCALE {
				// ! Currently we are only assuming 1 output per upscale request
				upscaleOutput, err = r.SetUpscaleSucceeded(msg.Input.ID, msg.Input.GenerationOutputID, msg.Outputs[0])
			} else {
				generationOutputs, err = r.SetGenerationSucceeded(msg.Input.ID, msg.Outputs)
			}
			if err != nil {
				klog.Errorf("--- Error setting %s succeeded for ID %s: %v", msg.Input.ProcessType, msg.Input.ID, err)
				return
			}
		}

	} else {
		klog.Warningf("--- Unknown webhook status, not sure how to proceed so not proceeding at all: %s", msg.Status)
		return
	}

	// Regardless of the status, we always send over sse so user knows what's up
	// Send message to user
	resp := responses.SSEStatusUpdateResponse{
		Status:    msg.Status,
		Id:        msg.Input.ID,
		StreamId:  streamIdStr,
		NSFWCount: msg.NSFWCount,
		Error:     cogErr,
	}
	// Upscale
	if msg.Status == responses.CogSucceeded && msg.Input.ProcessType == shared.UPSCALE {
		resp.Outputs = []responses.WebhookStatusUpdateOutputs{
			{
				ID:       upscaleOutput.ID,
				ImageUrl: upscaleOutput.ImageURL,
			},
		}
	} else if msg.Status == responses.CogSucceeded {
		// Generate or generate and upscale
		generateOutputs := make([]responses.WebhookStatusUpdateOutputs, len(generationOutputs))
		for i, output := range generationOutputs {
			generateOutputs[i] = responses.WebhookStatusUpdateOutputs{
				ID:               output.ID,
				ImageUrl:         output.ImageURL,
				UpscaledImageUrl: output.UpscaledImageURL,
				GalleryStatus:    output.GalleryStatus,
			}
		}
		resp.Outputs = generateOutputs
	}
	// Marshal
	respBytes, err := json.Marshal(resp)
	if err != nil {
		klog.Errorf("--- Error marshalling sse response: %v", err)
		return
	}

	// Broadcast to all clients subcribed to this stream
	r.Redis.Client.Publish(r.Redis.Client.Context(), shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes)
}
