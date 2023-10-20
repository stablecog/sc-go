// Description: Processes realtime messages from cog and updates the database
package repository

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/ent/voiceover"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Consider a generation/upscale a failure due to timeout
func (r *Repository) FailCogMessageDueToTimeoutIfTimedOut(msg requests.CogWebhookMessage) {
	deleted, err := r.Redis.DeleteCogRequestStreamID(r.Redis.Ctx, msg.Input.ID.String())
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
		var prefix string
		if msg.Input.ProcessType == shared.GENERATE || msg.Input.ProcessType == shared.GENERATE_AND_UPSCALE {
			prefix = "g"
		} else if msg.Input.ProcessType == shared.VOICEOVER {
			prefix = "v"
		} else {
			prefix = "u"
		}
		err := r.QueueThrottler.DecrementBy(1, fmt.Sprintf("%s:%s", prefix, msg.Input.UserID.String()))
		if err != nil {
			log.Error("Error decrementing queue count", "err", err, "user", msg.Input.UserID.String())
		}
	}

	// Remove from mq_log
	_, err = r.DeleteFromQueueLog(utils.Sha256(msg.Input.ID.String()), nil)
	if err != nil {
		log.Errorf("Error deleting from queue log: %v", err)
	}

	// ! Execute timeout failure
	// Get process type
	if msg.Input.ProcessType != shared.GENERATE && msg.Input.ProcessType != shared.UPSCALE && msg.Input.ProcessType != shared.GENERATE_AND_UPSCALE && msg.Input.ProcessType != shared.VOICEOVER {
		log.Error("Invalid process type from cog, can't handle message", "process_type", msg.Input.ProcessType)
		return
	}

	msg.Error = shared.TIMEOUT_ERROR

	// Only set for failures in case of refund
	var remainingCredits int

	var userId uuid.UUID
	if err := r.WithTx(func(tx *ent.Tx) error {
		db := tx.Client()
		if msg.Input.ProcessType == shared.UPSCALE {
			r.SetUpscaleFailed(msg.Input.ID.String(), msg.Error, db)
			user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
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
		} else if msg.Input.ProcessType == shared.VOICEOVER {
			r.SetVoiceoverFailed(msg.Input.ID.String(), msg.Error, db)
			user, err := r.DB.Voiceover.Query().Where(voiceover.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				log.Error("Error getting user ID from upscale", "err", err)
				return err
			}
			userId = user.ID
			creditAmount := utils.CalculateVoiceoverCredits(msg.Input.Prompt)
			success, err := r.RefundCreditsToUser(userId, creditAmount, db)
			if err != nil || !success {
				log.Error("Error refunding credits for voiceover", "user", userId.String(), "id", msg.Input.ID, "err", err)
				return err
			}
		} else {
			r.SetGenerationFailed(msg.Input.ID.String(), msg.Error, msg.NSFWCount, db)
			user, err := r.DB.Generation.Query().Where(generation.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
			if err != nil {
				log.Error("Error getting user ID from upscale", "err", err)
				return err
			}
			userId = user.ID
			success, err := r.RefundCreditsToUser(userId, *msg.Input.NumOutputs, db)
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
		Id:               msg.Input.ID.String(),
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
	_, err := r.Redis.DeleteCogRequestStreamID(r.Redis.Ctx, msg.Input.ID.String())
	if err != nil {
		log.Error("Error deleting stream ID from redis", "err", err)
	}

	// Get process type
	if msg.Input.ProcessType != shared.GENERATE && msg.Input.ProcessType != shared.UPSCALE && msg.Input.ProcessType != shared.GENERATE_AND_UPSCALE && msg.Input.ProcessType != shared.VOICEOVER {
		log.Error("Invalid process type from cog, can't handle message", "process_type", msg.Input.ProcessType)
		return fmt.Errorf("invalid process type from cog %s, can't handle message", msg.Input.ProcessType)
	}

	var upscaleOutput *ent.UpscaleOutput
	var voiceoverOutput *ent.VoiceoverOutput
	var generationOutputs []*ent.GenerationOutput
	var cogErr string

	// Remaining credits set only for failures
	var remainingCredits int

	// Handle started/failed/succeeded message types
	if msg.Status == requests.CogProcessing {
		// In goroutine since we want them to know it started asap
		if msg.Input.ProcessType == shared.UPSCALE {
			go r.SetUpscaleStarted(msg.Input.ID.String())
		} else if msg.Input.ProcessType == shared.VOICEOVER {
			go r.SetVoiceoverStarted(msg.Input.ID.String())
		} else {
			go r.SetGenerationStarted(msg.Input.ID.String())
		}
		_, err := r.DeleteFromQueueLog(utils.Sha256(msg.Input.ID.String()), nil)
		if err != nil {
			log.Errorf("Error deleting from queue log: %v", err)
		}
	} else if msg.Status == requests.CogFailed {
		// Delete from queue log
		_, err := r.DeleteFromQueueLog(utils.Sha256(msg.Input.ID.String()), nil)
		if err != nil {
			log.Errorf("Error deleting from queue log: %v", err)
		}
		// ! Failures for reasons other than NSFW,
		// ! We need to refund the credits
		if err := r.WithTx(func(tx *ent.Tx) error {
			db := tx.Client()
			var userId uuid.UUID
			if msg.Input.ProcessType == shared.UPSCALE {
				r.SetUpscaleFailed(msg.Input.ID.String(), msg.Error, db)
				user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
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
			} else if msg.Input.ProcessType == shared.VOICEOVER {
				r.SetVoiceoverFailed(msg.Input.ID.String(), msg.Error, db)
				user, err := r.DB.Voiceover.Query().Where(voiceover.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
				if err != nil {
					log.Error("Error getting user ID from upscale", "err", err)
					return err
				}
				userId = user.ID
				creditAmount := utils.CalculateVoiceoverCredits(msg.Input.Prompt)
				success, err := r.RefundCreditsToUser(userId, creditAmount, db)
				if err != nil || !success {
					log.Error("Error refunding credits for voiceover", "user", userId.String(), "id", msg.Input.ID, "err", err)
					return err
				}
			} else {
				r.SetGenerationFailed(msg.Input.ID.String(), msg.Error, msg.NSFWCount, db)
				user, err := r.DB.Generation.Query().Where(generation.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
				if err != nil {
					log.Error("Error getting user ID from upscale", "err", err)
					return err
				}
				userId = user.ID
				success, err := r.RefundCreditsToUser(userId, *msg.Input.NumOutputs, db)
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
		_, err := r.DeleteFromQueueLog(utils.Sha256(msg.Input.ID.String()), nil)
		if err != nil {
			log.Errorf("Error deleting from queue log: %v", err)
		}
		if len(msg.Output.Images) == 0 && msg.Input.ProcessType != shared.VOICEOVER {
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
					err := r.SetUpscaleFailed(msg.Input.ID.String(), cogErr, db)
					if err != nil {
						log.Error("Error setting upscale failed", "err", err)
						return err
					}
					if processRefund {
						user, err := r.DB.Upscale.Query().Where(upscale.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
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
					err := r.SetGenerationFailed(msg.Input.ID.String(), cogErr, msg.NSFWCount, db)
					if err != nil {
						log.Error("Error setting generation failed", "err", err)
						return err
					}
					if processRefund {
						user, err := r.DB.Generation.Query().Where(generation.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
						if err != nil {
							log.Error("Error getting user ID from upscale", "err", err)
							return err
						}
						success, err := r.RefundCreditsToUser(user.ID, *msg.Input.NumOutputs, db)
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
		} else if msg.Input.ProcessType == shared.VOICEOVER && len(msg.Output.AudioFiles) == 0 {
			if err := r.WithTx(func(tx *ent.Tx) error {
				db := tx.Client()
				err := r.SetVoiceoverFailed(msg.Input.ID.String(), cogErr, db)
				if err != nil {
					log.Error("Error setting voiceover failed", "err", err)
					return err
				}
				user, err := r.DB.Voiceover.Query().Where(voiceover.IDEQ(msg.Input.ID)).QueryUser().Select(user.FieldID).First(r.Ctx)
				if err != nil {
					log.Error("Error getting user ID from voiceover", "err", err)
					return err
				}
				creditAmount := utils.CalculateVoiceoverCredits(msg.Input.Prompt)
				success, err := r.RefundCreditsToUser(user.ID, creditAmount, db)
				if err != nil || !success {
					log.Error("Error refunding credits for voiceover", "user", user.ID.String(), "id", msg.Input.ID, "err", err)
					return err
				}
				remainingCredits, err = r.GetNonExpiredCreditTotalForUser(user.ID, db)
				if err != nil {
					log.Error("Error getting remaining credits", "err", err)
					return err
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
				upscaleOutput, err = r.SetUpscaleSucceeded(msg.Input.ID.String(), msg.Input.GenerationOutputID, msg.Input.Image, msg.Output)
			} else if msg.Input.ProcessType == shared.VOICEOVER {
				voiceoverOutput, err = r.SetVoiceoverSucceeded(msg.Input.ID.String(), msg.Input.Prompt, msg.Input.SubmitToGallery, msg.Output)
			} else {
				generationOutputs, err = r.SetGenerationSucceeded(msg.Input.ID.String(), msg.Input.OriginalPrompt, msg.Input.OriginalNegativePrompt, msg.Input.SubmitToGallery, msg.Output, msg.NSFWCount)
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
	var respBytes []byte
	if msg.Input.ProcessType != shared.VOICEOVER {
		resp := TaskStatusUpdateResponse{
			Status:           msg.Status,
			Id:               msg.Input.ID.String(),
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
					WasAutoSubmitted: msg.Input.WasAutoSubmitted,
					IsPublic:         msg.Input.WasAutoSubmitted,
				}
			}
			resp.Outputs = generateOutputs
		}
		// Marshal
		respBytes, err = json.Marshal(resp)
		if err != nil {
			log.Error("Error marshalling sse response", "err", err)
			return err
		}
	} else {
		// Voiceover
		resp := TaskStatusUpdateResponse{
			Status:           msg.Status,
			Id:               msg.Input.ID.String(),
			UIId:             msg.Input.UIId,
			StreamId:         msg.Input.StreamID,
			Error:            cogErr,
			ProcessType:      msg.Input.ProcessType,
			RemainingCredits: remainingCredits,
		}
		if msg.Status == requests.CogSucceeded {
			audioFileURL := utils.GetURLFromAudioFilePath(voiceoverOutput.AudioPath)
			resp.Outputs = []GenerationUpscaleOutput{
				{
					ID:               voiceoverOutput.ID,
					AudioFileUrl:     audioFileURL,
					AudioDuration:    &voiceoverOutput.AudioDuration,
					WasAutoSubmitted: msg.Input.WasAutoSubmitted,
					IsPublic:         msg.Input.WasAutoSubmitted,
				},
			}
		}
		// Marshal
		respBytes, err = json.Marshal(resp)
		if err != nil {
			log.Error("Error marshalling sse response", "err", err)
			return err
		}
	}

	// Dec queue count
	if msg.Input.UserID != nil && (msg.Status == requests.CogSucceeded || msg.Status == requests.CogFailed) {
		var prefix string
		if msg.Input.ProcessType == shared.GENERATE || msg.Input.ProcessType == shared.GENERATE_AND_UPSCALE {
			prefix = "g"
		} else if msg.Input.ProcessType == shared.VOICEOVER {
			prefix = "v"
		} else {
			prefix = "u"
		}
		err := r.QueueThrottler.DecrementBy(1, fmt.Sprintf("%s:%s", prefix, msg.Input.UserID.String()))
		if err != nil {
			log.Error("Error decrementing queue count", "err", err, "user", msg.Input.UserID.String())
		}
	}

	// Broadcast to all clients subcribed to this stream
	r.Redis.Client.Publish(r.Redis.Ctx, shared.REDIS_SSE_BROADCAST_CHANNEL, respBytes)
	return nil
}
