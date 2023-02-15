package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

func (r *Repository) GetUserByStripeCustomerId(customerId string) (*ent.User, error) {
	user, err := r.DB.User.Query().Where(user.StripeCustomerIDEQ(customerId)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		klog.Errorf("Error getting user by stripe customer ID: %v", err)
		return nil, err
	}
	return user, nil
}

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

	inputUuid, err := uuid.Parse(msg.Input.ID)
	if err != nil {
		klog.Errorf("--- Error parsing input ID, not a UUID: %v", err)
		return
	}

	// Handle started/failed/succeeded message types
	if msg.Status == responses.CogProcessing {
		// In goroutine since we want them to know it started asap
		if msg.Input.ProcessType == shared.UPSCALE {
			go r.SetUpscaleStarted(msg.Input.ID)
		} else {
			go r.SetGenerationStarted(msg.Input.ID)
		}
	} else if msg.Status == responses.CogFailed {
		// ! Failures for reasons other than NSFW,
		// ! We need to refund the credits
		// TODO - we need to handle what to do if the refund fails for whatever reason
		var userId uuid.UUID
		if msg.Input.ProcessType == shared.UPSCALE {
			r.SetUpscaleFailed(msg.Input.ID, msg.Error)
			user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				klog.Errorf("--- Error getting user ID from upscale: %v", err)
				return
			}
			userId = user.ID
			// Upscale is always 1 credit
			success, err := r.RefundCreditsToUser(userId, 1)
			if err != nil || !success {
				klog.Errorf("--- Error refunding credits to user %s for upscale %s: %v", userId.String(), msg.Input.ID, err)
				return
			}
		} else {
			r.SetGenerationFailed(msg.Input.ID, msg.Error, msg.NSFWCount)
			user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				klog.Errorf("--- Error getting user ID from upscale: %v", err)
				return
			}
			userId = user.ID
			// Generation credits is num_outputs
			numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
			if err != nil {
				klog.Errorf("--- Error parsing num outputs: %v", err)
				return
			}
			success, err := r.RefundCreditsToUser(userId, int32(numOutputs))
			if err != nil || !success {
				klog.Errorf("--- Error refunding credits to user %s for generation %s: %v", userId.String(), msg.Input.ID, err)
				return
			}
		}
		cogErr = msg.Error
	} else if msg.Status == responses.CogSucceeded {
		if len(msg.Outputs) == 0 {
			// NSFW comes back as a success, but with no outputs and nsfw count
			if msg.NSFWCount > 0 {
				cogErr = "NSFW"
			} else {
				cogErr = "No outputs"
				// TODO - this should probably be refunded if it ever happens
			}
			if msg.Input.ProcessType == shared.UPSCALE {
				r.SetUpscaleFailed(msg.Input.ID, cogErr)
			} else {
				r.SetGenerationFailed(msg.Input.ID, cogErr, msg.NSFWCount)
			}
			msg.Status = responses.CogFailed
		} else {
			if msg.Input.ProcessType == shared.UPSCALE {
				// ! Currently we are only assuming 1 output per upscale request
				upscaleOutput, err = r.SetUpscaleSucceeded(msg.Input.ID, msg.Input.GenerationOutputID, msg.Outputs[0])
			} else {
				generationOutputs, err = r.SetGenerationSucceeded(msg.Input.ID, msg.Outputs, msg.NSFWCount)
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
		imageUrl, err := utils.ParseS3UrlToURL(upscaleOutput.ImagePath)
		if err != nil {
			klog.Errorf("--- Error parsing s3 url to url: %v", err)
			imageUrl = upscaleOutput.ImagePath
		}
		resp.Outputs = []responses.GenerationOutputResponse{
			{
				ID:       upscaleOutput.ID,
				ImageUrl: imageUrl,
			},
		}
	} else if msg.Status == responses.CogSucceeded {
		// Generate or generate and upscale
		generateOutputs := make([]responses.GenerationOutputResponse, len(generationOutputs))
		for i, output := range generationOutputs {
			// Parse S3 URLs to usable URLs
			imageUrl, err := utils.ParseS3UrlToURL(output.ImagePath)
			if err != nil {
				klog.Errorf("Error parsing image url %s: %v", output.ImagePath, err)
				imageUrl = output.ImagePath
			}
			var upscaledImageUrl string
			if output.UpscaledImagePath != nil {
				upscaledImageUrl, err = utils.ParseS3UrlToURL(*output.UpscaledImagePath)
				if err != nil {
					klog.Errorf("Error parsing upscaled image url %s: %v", *output.UpscaledImagePath, err)
					upscaledImageUrl = *output.UpscaledImagePath
				}
			}
			generateOutputs[i] = responses.GenerationOutputResponse{
				ID:               output.ID,
				ImageUrl:         imageUrl,
				UpscaledImageUrl: upscaledImageUrl,
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
