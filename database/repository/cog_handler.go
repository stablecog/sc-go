// Description: Processes realtime messages from cog and updates the database
package repository

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Consider a generation/upscale a failure due to timeout
func (r *Repository) FailCogMessageDueToTimeoutIfTimedOut(msg requests.CogWebhookMessage) {
	deleted, err := r.Redis.DeleteCogRequestStreamID(r.Redis.Ctx, msg.Input.ID)
	if err != nil {
		log.Error("Error deleting stream ID from redis", "err", err)
		return
	}
	if deleted == 0 {
		// Means it didnt time out
		return
	}

	// Dec queue count
	if msg.Input.UserID != nil {
		// Parse num
		numOutputs := 1
		if msg.Input.ProcessType != shared.UPSCALE {
			// Parse as int
			numOutputs, err = strconv.Atoi(msg.Input.NumOutputs)
			if err != nil {
				log.Error("Error parsing num outputs", "err", err)
			}
		}
		err := r.QueueThrottler.DecrementBy(numOutputs, msg.Input.UserID.String())
		if err != nil {
			log.Error("Error decrementing queue count", "err", err, "user", msg.Input.UserID.String())
		}
	}

	// ! Execute timeout failure
	// Get process type
	if msg.Input.ProcessType != shared.GENERATE && msg.Input.ProcessType != shared.UPSCALE && msg.Input.ProcessType != shared.GENERATE_AND_UPSCALE {
		log.Error("Invalid process type from cog, can't handle message", "process_type", msg.Input.ProcessType)
		return
	}

	msg.Error = shared.TIMEOUT_ERROR

	inputUuid, err := uuid.Parse(msg.Input.ID)
	if err != nil {
		log.Error("Error parsing input ID, not a UUID", "err", err)
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
				log.Error("Error getting user ID from upscale", "err", err)
				return err
			}
			userId = user.ID
			// Upscale is always 1 credit
			success, err := r.RefundCreditsToUser(userId, 1, db)
			if err != nil || !success {
				log.Error("Error refunding credits for upscale", "user", userId.String(), "id", msg.Input.ID, "err", err)
				return err
			}
		} else {
			r.SetGenerationFailed(msg.Input.ID, msg.Error, msg.NSFWCount, db)
			user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				log.Error("Error getting user ID from upscale", "err", err)
				return err
			}
			userId = user.ID
			// Generation credits is num_outputs
			numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
			if err != nil {
				log.Error("Error parsing num outputs", "err", err)
				return err
			}
			success, err := r.RefundCreditsToUser(userId, int32(numOutputs), db)
			if err != nil || !success {
				log.Error("Error refunding credits for generation", "user", userId.String(), "id", msg.Input.ID, "err", err)
				return err
			}
		}

		remainingCredits, err = r.GetNonExpiredCreditTotalForUser(userId, db)
		if err != nil {
			log.Error("Error getting remaining credits", "err", err)
			return err
		}

		return nil
	}); err != nil {
		log.Error("Error in processing timeout failure transaction", "err", err)
		return
	}

	// Regardless of the status, we always send over sse so user knows what's up
	// Send message to user
	resp := TaskStatusUpdateResponse{
		Status:           msg.Status,
		Id:               msg.Input.ID,
		UIId:             msg.Input.UIId,
		StreamId:         msg.Input.StreamID,
		NSFWCount:        msg.NSFWCount,
		Error:            msg.Error,
		ProcessType:      msg.Input.ProcessType,
		RemainingCredits: remainingCredits,
	}

	// Marshal
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Error("Error marshalling sse response", "err", err)
		return
	}

	// Broadcast to all clients subcribed to this stream
	r.Redis.Client.Publish(r.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes)
}

// Process a cog message into database
func (r *Repository) ProcessCogMessage(msg requests.CogWebhookMessage) error {
	// Delete timeout key
	_, err := r.Redis.DeleteCogRequestStreamID(r.Redis.Ctx, msg.Input.ID)
	if err != nil {
		log.Error("Error deleting stream ID from redis", "err", err)
	}

	// Get process type
	if msg.Input.ProcessType != shared.GENERATE && msg.Input.ProcessType != shared.UPSCALE && msg.Input.ProcessType != shared.GENERATE_AND_UPSCALE {
		log.Error("Invalid process type from cog, can't handle message", "process_type", msg.Input.ProcessType)
		return fmt.Errorf("invalid process type from cog %s, can't handle message", msg.Input.ProcessType)
	}

	var upscaleOutput *ent.UpscaleOutput
	var generationOutputs []*ent.GenerationOutput
	var cogErr string

	inputUuid, err := uuid.Parse(msg.Input.ID)
	if err != nil {
		log.Error("Error parsing input ID, not a UUID", "err", err)
		return err
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
					log.Error("Error getting user ID from upscale", "err", err)
					return err
				}
				userId = user.ID
				// Upscale is always 1 credit
				success, err := r.RefundCreditsToUser(userId, 1, db)
				if err != nil || !success {
					log.Error("Error refunding credits for upscale", "user", userId.String(), "id", msg.Input.ID, "err", err)
					return err
				}
			} else {
				r.SetGenerationFailed(msg.Input.ID, msg.Error, msg.NSFWCount, db)
				user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
				if err != nil {
					log.Error("Error getting user ID from upscale", "err", err)
					return err
				}
				userId = user.ID
				// Generation credits is num_outputs
				numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
				if err != nil {
					log.Error("Error parsing num outputs", "err", err)
					return err
				}
				success, err := r.RefundCreditsToUser(userId, int32(numOutputs), db)
				if err != nil || !success {
					log.Error("Error refunding credits for generation", "user", userId.String(), "id", msg.Input.ID, "err", err)
					return err
				}
			}
			remainingCredits, err = r.GetNonExpiredCreditTotalForUser(userId, db)
			if err != nil {
				log.Error("Error getting remaining credits", "err", err)
				return err
			}
			cogErr = msg.Error
			return nil
		}); err != nil {
			log.Error("Error with transaction in cog message process", "err", err)
			return err
		}
	} else if msg.Status == requests.CogSucceeded {
		if len(msg.Output.Images) == 0 {
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
						log.Error("Error setting upscale failed", "err", err)
						return err
					}
					if processRefund {
						user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
						if err != nil {
							log.Error("Error getting user ID from upscale", "err", err)
							return err
						}
						success, err := r.RefundCreditsToUser(user.ID, 1, db)
						if err != nil || !success {
							log.Error("Error refunding credits for upscale", "user", user.ID.String(), "id", msg.Input.ID, "err", err)
							return err
						}
						remainingCredits, err = r.GetNonExpiredCreditTotalForUser(user.ID, db)
						if err != nil {
							log.Error("Error getting remaining credits", "err", err)
							return err
						}
					}
				} else {
					err := r.SetGenerationFailed(msg.Input.ID, cogErr, msg.NSFWCount, db)
					if err != nil {
						log.Error("Error setting generation failed", "err", err)
						return err
					}
					if processRefund {
						user, err := r.DB.Generation.Query().Where(generation.IDEQ(inputUuid)).QueryUser().Select(user.FieldID).First(r.Ctx)
						if err != nil {
							log.Error("Error getting user ID from upscale", "err", err)
							return err
						}
						numOutputs, err := strconv.Atoi(msg.Input.NumOutputs)
						if err != nil {
							log.Error("Error parsing num outputs", "err", err)
							return err
						}
						success, err := r.RefundCreditsToUser(user.ID, int32(numOutputs), db)
						if err != nil || !success {
							log.Error("Error refunding credits for upscale", "user", user.ID.String(), "id", msg.Input.ID, "err", err)
							return err
						}
						remainingCredits, err = r.GetNonExpiredCreditTotalForUser(user.ID, db)
						if err != nil {
							log.Error("Error getting remaining credits", "err", err)
							return err
						}
					}
				}
				msg.Status = requests.CogFailed
				return nil
			}); err != nil {
				log.Error("Error with transaction in cog message process", "err", err)
				return err
			}
		} else {
			if msg.Input.ProcessType == shared.UPSCALE {
				// ! Currently we are only assuming 1 output per upscale request
				upscaleOutput, err = r.SetUpscaleSucceeded(msg.Input.ID, msg.Input.GenerationOutputID, msg.Input.Image, msg.Output)
			} else {
				generationOutputs, err = r.SetGenerationSucceeded(msg.Input.ID, msg.Input.Prompt, msg.Input.NegativePrompt, msg.Output, msg.NSFWCount)
			}
			if err != nil {
				log.Error("Error setting process succeeded", "process_type", msg.Input.ProcessType, "id", msg.Input.ID, "err", err)
				return err
			}
		}

	} else {
		log.Warn("Unknown webhook status, not sure how to proceed so not proceeding at all", "status", msg.Status)
		return fmt.Errorf("invalid status %s", msg.Status)
	}

	// Regardless of the status, we always send over sse so user knows what's up
	// Send message to user
	resp := TaskStatusUpdateResponse{
		Status:           msg.Status,
		Id:               msg.Input.ID,
		UIId:             msg.Input.UIId,
		StreamId:         msg.Input.StreamID,
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
				ID:           upscaleOutput.ID,
				ImageUrl:     imageUrl,
				InitImageUrl: msg.Input.Image,
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
		log.Error("Error marshalling sse response", "err", err)
		return err
	}

	// Dec queue count
	if msg.Input.UserID != nil && (msg.Status == requests.CogSucceeded || msg.Status == requests.CogFailed) {
		numOutputs := 1
		if msg.Input.ProcessType != shared.UPSCALE {
			// Parse as int
			numOutputs, err = strconv.Atoi(msg.Input.NumOutputs)
			if err != nil {
				log.Error("Error parsing num outputs", "err", err)
			}
		}
		err := r.QueueThrottler.DecrementBy(numOutputs, msg.Input.UserID.String())
		if err != nil {
			log.Error("Error decrementing queue count", "err", err, "user", msg.Input.UserID.String())
		}
	}

	// Broadcast to all clients subcribed to this stream
	r.Redis.Client.Publish(r.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes)
	return nil
}
