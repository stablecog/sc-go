package sse

import (
	"encoding/json"

	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

// Broadcasts message from sc-worker to client(s) SSE stream(s)
func (h *Hub) BroadcastWorkerMessageToClient(msg responses.CogStatusUpdate) {
	// See if this request belongs to our instance
	// We may be scaled, and the request may have been created by another instance - so if we don't have it here we ignore it
	streamIdStr := h.CogRequestSSEConnMap.Get(msg.Input.ID)
	if streamIdStr == "" {
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
	var err error

	// Handle started/failed/succeeded message types
	if msg.Status == responses.CogProcessing {
		// In goroutine since we want them to know it started asap
		if msg.Input.ProcessType == shared.UPSCALE {
			go h.Repo.SetUpscaleStarted(msg.Input.ID)
		} else {
			go h.Repo.SetGenerationStarted(msg.Input.ID)
		}
	} else if msg.Status == responses.CogFailed {
		if msg.Input.ProcessType == shared.UPSCALE {
			h.Repo.SetUpscaleFailed(msg.Input.ID, msg.Error)
		} else {
			h.Repo.SetGenerationFailed(msg.Input.ID, msg.Error)
		}
		cogErr = msg.Error
	} else if msg.Status == responses.CogSucceeded {
		if len(msg.Outputs) == 0 {
			cogErr = "No outputs"
			if msg.Input.ProcessType == shared.UPSCALE {
				h.Repo.SetUpscaleFailed(msg.Input.ID, cogErr)
			} else {
				h.Repo.SetGenerationFailed(msg.Input.ID, cogErr)
			}
			msg.Status = responses.CogFailed
		} else {
			if msg.Input.ProcessType == shared.UPSCALE {
				// ! Currently we are only assuming 1 output per upscale request
				upscaleOutput, err = h.Repo.SetUpscaleSucceeded(msg.Input.ID, msg.Input.GenerationOutputID, msg.Outputs[0])
			} else {
				generationOutputs, err = h.Repo.SetGenerationSucceeded(msg.Input.ID, msg.Outputs)
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
	h.BroadcastToClientsWithUid(streamIdStr, respBytes)
}
