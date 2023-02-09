package sse

import (
	"encoding/json"

	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/utils"
	"k8s.io/klog/v2"
)

// TODO - we can de-dupe a lot of this logic

// Broadcasts an upscale message from sc-worker to relevant client
func (h *Hub) BroadcastUpscaleMessage(msg responses.CogStatusUpdate) {
	// See if this request belongs to our instance
	streamIdStr := h.CogRequestSSEConnMap.Get(msg.Input.Id)
	if streamIdStr == "" {
		return
	}
	// Ensure is valid
	if !utils.IsSha256Hash(streamIdStr) {
		// Not sure how we get here, we never should
		klog.Errorf("--- Invalid SSE stream id: %s", streamIdStr)
		return
	}

	// Processing, update this upscale as started in database
	var output *ent.UpscaleOutput
	var cogErr string
	var err error
	if msg.Status == responses.CogProcessing {
		// In goroutine since we want them to know it started asap
		go h.Repo.SetUpscaleStarted(msg.Input.Id)
	} else if msg.Status == responses.CogFailed {
		h.Repo.SetUpscaleFailed(msg.Input.Id, msg.Error)
		cogErr = msg.Error
	} else if msg.Status == responses.CogSucceeded {
		if len(msg.Outputs) == 0 {
			cogErr = "No outputs"
			h.Repo.SetUpscaleFailed(msg.Input.Id, cogErr)
			msg.Status = responses.CogFailed
		} else {
			output, err = h.Repo.SetUpscaleSucceeded(msg.Input.Id, msg.Input.GenerationOutputID, msg.Outputs[0])
			if err != nil {
				klog.Errorf("--- Error setting upscale succeeded for ID %s: %v", msg.Input.Id, err)
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
		Id:        msg.Input.Id,
		NSFWCount: msg.NSFWCount,
		Error:     cogErr,
	}
	if msg.Status == responses.CogSucceeded {
		resp.Outputs = []responses.WebhookStatusUpdateOutputs{
			{
				ID:       output.ID,
				ImageUrl: output.ImageURL,
			},
		}
	}
	// Marshal
	respBytes, err := json.Marshal(resp)
	if err != nil {
		klog.Errorf("--- Error marshalling sse started response: %v", err)
		return
	}
	h.BroadcastToClientsWithUid(streamIdStr, respBytes)
}

// Broadcast a generate message from sc-worker to relevant client
func (h *Hub) BroadcastGenerateMessage(msg responses.CogStatusUpdate) {
	// See if this request belongs to our instance
	streamIdStr := h.CogRequestSSEConnMap.Get(msg.Input.Id)
	if streamIdStr == "" {
		return
	}
	// Ensure is valid
	if !utils.IsSha256Hash(streamIdStr) {
		// Not sure how we get here, we never should
		klog.Errorf("--- Invalid sse stream id: %s", streamIdStr)
		return
	}

	// Processing, update this generation as started in database
	var outputs []*ent.GenerationOutput
	var cogErr string
	var err error
	if msg.Status == responses.CogProcessing {
		// In goroutine since we want them to know it started asap
		go h.Repo.SetGenerationStarted(msg.Input.Id)
	} else if msg.Status == responses.CogFailed {
		h.Repo.SetGenerationFailed(msg.Input.Id, msg.Error)
		cogErr = msg.Error
	} else if msg.Status == responses.CogSucceeded {
		// NSFW counts as failure
		if len(msg.Outputs) == 0 && msg.NSFWCount > 0 {
			cogErr = "NSFW"
			h.Repo.SetGenerationFailed(msg.Input.Id, cogErr)
			msg.Status = responses.CogFailed
		} else if len(msg.Outputs) == 0 && msg.NSFWCount == 0 {
			cogErr = "No outputs"
			h.Repo.SetGenerationFailed(msg.Input.Id, cogErr)
			msg.Status = responses.CogFailed
		} else {
			outputs, err = h.Repo.SetGenerationSucceeded(msg.Input.Id, msg.Outputs)
			if err != nil {
				klog.Errorf("--- Error setting generation succeeded for ID %s: %v", msg.Input.Id, err)
				return
			}
		}
	} else {
		klog.Warningf("--- Unknown webhook status, not sure how to proceed so not proceeding at all: %s", msg.Status)
		return
	}

	// Regardless of the status, we always send over SSE so user knows what's up
	// Send message to user
	resp := responses.SSEStatusUpdateResponse{
		Status:    msg.Status,
		Id:        msg.Input.Id,
		NSFWCount: msg.NSFWCount,
		Error:     cogErr,
	}

	if msg.Status == responses.CogSucceeded {
		generateOutputs := make([]responses.WebhookStatusUpdateOutputs, len(outputs))
		for i, output := range outputs {
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
		klog.Errorf("--- Error marshalling sse started response: %v", err)
		return
	}
	h.BroadcastToClientsWithUid(streamIdStr, respBytes)
}
