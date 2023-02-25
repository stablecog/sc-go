// Description: Processes realtime messages from cog and updates the database
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
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

// Consider a generation/upscale a failure due to timeout
func (r *Repository) FailCogMessageDueToTimeoutIfTimedOut(msg requests.CogRedisMessage) {
	redisKey := fmt.Sprintf("second:%s", msg.Input.ID)
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
		// Means we don't need to timeout
		return
	}

	// ! Execute timeout failure

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

	msg.Error = shared.TIMEOUT_ERROR

	inputUuid, err := uuid.Parse(msg.Input.ID)
	if err != nil {
		klog.Errorf("--- Error parsing input ID, not a UUID: %v", err)
		return
	}

	// Only set for failures in case of refund
	var remainingCredits int

	var userId uuid.UUID
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()
		if msg.Input.ProcessType == shared.UPSCALE {
			r.SetUpscaleFailed(msg.Input.ID, msg.Error, db)
			user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				klog.Errorf("--- Error getting user ID from upscale: %v", err)
				return err
			}
			userId = user.ID
			// Upscale is always 1 credit
			success, err := r.RefundCreditsToUser(userId, 1, db)
			if err != nil || !success {
				klog.Errorf("--- Error refunding credits to user %s for upscale %s: %v", userId.String(), msg.Input.ID, err)
				return err
			}
		} else {
			r.SetGenerationFailed(msg.Input.ID, msg.Error, msg.NSFWCount, db)
			user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				klog.Errorf("--- Error getting user ID from upscale: %v", err)
				return err
			}
			userId = user.ID
			// Generation credits is num_outputs
			numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
			if err != nil {
				klog.Errorf("--- Error parsing num outputs: %v", err)
				return err
			}
			success, err := r.RefundCreditsToUser(userId, int32(numOutputs), db)
			if err != nil || !success {
				klog.Errorf("--- Error refunding credits to user %s for generation %s: %v", userId.String(), msg.Input.ID, err)
				return err
			}
		}

		remainingCredits, err = r.GetNonExpiredCreditTotalForUser(userId, db)
		if err != nil {
			klog.Errorf("Error getting remaining credits: %v", err)
			return err
		}

		return nil
	}); err != nil {
		klog.Errorf("--- Error in processing timeout failure transaction: %v", err)
		return
	}

	// Regardless of the status, we always send over sse so user knows what's up
	// Send message to user
	resp := TaskStatusUpdateResponse{
		Status:           msg.Status,
		Id:               msg.Input.ID,
		StreamId:         streamIdStr,
		NSFWCount:        msg.NSFWCount,
		Error:            msg.Error,
		ProcessType:      msg.Input.ProcessType,
		RemainingCredits: remainingCredits,
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

// Process a cog message into database
func (r *Repository) ProcessCogMessage(msg requests.CogRedisMessage) {
	redisKey := fmt.Sprintf("first:%s", msg.Input.ID)
	if msg.Status != requests.CogProcessing {
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

	// Remaining credits set only for failures
	var remainingCredits int

	// Handle started/failed/succeeded message types
	if msg.Status == requests.CogProcessing {
		// In goroutine since we want them to know it started asap
		if msg.Input.ProcessType == shared.UPSCALE {
			go r.SetUpscaleStarted(msg.Input.ID)
		} else {
			go r.SetGenerationStarted(msg.Input.ID)
		}
	} else if msg.Status == requests.CogFailed {
		// ! Failures for reasons other than NSFW,
		// ! We need to refund the credits
		if err := r.WithTx(func(tx *ent.Tx) error {
			db := tx.Client()
			var userId uuid.UUID
			if msg.Input.ProcessType == shared.UPSCALE {
				r.SetUpscaleFailed(msg.Input.ID, msg.Error, db)
				user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
				if err != nil {
					klog.Errorf("--- Error getting user ID from upscale: %v", err)
					return err
				}
				userId = user.ID
				// Upscale is always 1 credit
				success, err := r.RefundCreditsToUser(userId, 1, db)
				if err != nil || !success {
					klog.Errorf("--- Error refunding credits to user %s for upscale %s: %v", userId.String(), msg.Input.ID, err)
					return err
				}
			} else {
				r.SetGenerationFailed(msg.Input.ID, msg.Error, msg.NSFWCount, db)
				user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
				if err != nil {
					klog.Errorf("--- Error getting user ID from upscale: %v", err)
					return err
				}
				userId = user.ID
				// Generation credits is num_outputs
				numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
				if err != nil {
					klog.Errorf("--- Error parsing num outputs: %v", err)
					return err
				}
				success, err := r.RefundCreditsToUser(userId, int32(numOutputs), db)
				if err != nil || !success {
					klog.Errorf("--- Error refunding credits to user %s for generation %s: %v", userId.String(), msg.Input.ID, err)
					return err
				}
			}
			remainingCredits, err = r.GetNonExpiredCreditTotalForUser(userId, db)
			if err != nil {
				klog.Errorf("Error getting remaining credits: %v", err)
				return err
			}
			cogErr = msg.Error
			return nil
		}); err != nil {
			klog.Errorf("--- Error with transaction in cog message process: %v", err)
			return
		}
	} else if msg.Status == requests.CogSucceeded {
		if len(msg.Outputs) == 0 {
			if err := r.WithTx(func(tx *ent.Tx) error {
				db := tx.Client()
				// NSFW comes back as a success, but with no outputs and nsfw count
				processRefund := false
				if msg.NSFWCount > 0 {
					cogErr = shared.NSFW_ERROR
				} else {
					cogErr = "No outputs"
					processRefund = true
				}
				if msg.Input.ProcessType == shared.UPSCALE {
					err := r.SetUpscaleFailed(msg.Input.ID, cogErr, db)
					if err != nil {
						klog.Errorf("--- Error setting upscale failed: %v", err)
						return err
					}
					if processRefund {
						user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
						if err != nil {
							klog.Errorf("--- Error getting user ID from upscale: %v", err)
							return err
						}
						success, err := r.RefundCreditsToUser(user.ID, 1, db)
						if err != nil || !success {
							klog.Errorf("--- Error refunding credits to user %s for upscale %s: %v", user.ID.String(), msg.Input.ID, err)
							return err
						}
						remainingCredits, err = r.GetNonExpiredCreditTotalForUser(user.ID, db)
						if err != nil {
							klog.Errorf("Error getting remaining credits: %v", err)
							return err
						}
					}
				} else {
					err := r.SetGenerationFailed(msg.Input.ID, cogErr, msg.NSFWCount, db)
					if err != nil {
						klog.Errorf("--- Error setting generation failed: %v", err)
						return err
					}
					if processRefund {
						user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
						if err != nil {
							klog.Errorf("--- Error getting user ID from upscale: %v", err)
							return err
						}
						numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
						if err != nil {
							klog.Errorf("--- Error parsing num outputs: %v", err)
							return err
						}
						success, err := r.RefundCreditsToUser(user.ID, int32(numOutputs), db)
						if err != nil || !success {
							klog.Errorf("--- Error refunding credits to user %s for generation %s: %v", user.ID.String(), msg.Input.ID, err)
							return err
						}
						remainingCredits, err = r.GetNonExpiredCreditTotalForUser(user.ID, db)
						if err != nil {
							klog.Errorf("Error getting remaining credits: %v", err)
							return err
						}
					}
				}
				msg.Status = requests.CogFailed
				return nil
			}); err != nil {
				klog.Errorf("--- Error with transaction in cog message process: %v", err)
				return
			}
		} else {
			if msg.Input.ProcessType == shared.UPSCALE {
				// ! Currently we are only assuming 1 output per upscale request
				upscaleOutput, err = r.SetUpscaleSucceeded(msg.Input.ID, msg.Input.GenerationOutputID, msg.Input.Image, msg.Outputs[0])
			} else {
				generationOutputs, err = r.SetGenerationSucceeded(msg.Input.ID, msg.Input.Prompt, msg.Input.NegativePrompt, msg.Outputs, msg.NSFWCount)
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
	resp := TaskStatusUpdateResponse{
		Status:           msg.Status,
		Id:               msg.Input.ID,
		StreamId:         streamIdStr,
		NSFWCount:        msg.NSFWCount,
		Error:            cogErr,
		ProcessType:      msg.Input.ProcessType,
		RemainingCredits: remainingCredits,
	}
	// Upscale
	if msg.Status == requests.CogSucceeded && msg.Input.ProcessType == shared.UPSCALE {
		imageUrl := utils.GetURLFromImagePath(upscaleOutput.ImagePath)
		resp.Outputs = []GenerationUpscaleOutput{
			{
				ID:            upscaleOutput.ID,
				ImageUrl:      imageUrl,
				InputImageUrl: msg.Input.Image,
			},
		}
		outputId, err := uuid.Parse(msg.Input.GenerationOutputID)
		if err == nil {
			resp.Outputs[0].OutputID = &outputId
		}
	} else if msg.Status == requests.CogSucceeded {
		// Generate or generate and upscale
		generateOutputs := make([]GenerationUpscaleOutput, len(generationOutputs))
		for i, output := range generationOutputs {
			// Parse S3 URLs to usable URLs
			imageUrl := utils.GetURLFromImagePath(output.ImagePath)
			var upscaledImageUrl string
			if output.UpscaledImagePath != nil {
				upscaledImageUrl = utils.GetURLFromImagePath(*output.UpscaledImagePath)
			}
			generateOutputs[i] = GenerationUpscaleOutput{
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
