package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stretchr/testify/assert"
)

func TestHandleReviewGallerySubmission(t *testing.T) {
	// ! Can approve generation
	var targetGUid uuid.UUID
	// Find goutput not approved
	goutput, err := MockController.Repo.DB.GenerationOutput.Query().Where(generationoutput.GalleryStatusNEQ(generationoutput.GalleryStatusAccepted)).First(MockController.Repo.Ctx)
	assert.Nil(t, err)
	targetGUid = goutput.ID

	reqBody := requests.ReviewGalleryRequest{
		Action:              requests.GalleryApproveAction,
		GenerationOutputIDs: []uuid.UUID{targetGUid},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var reviewResp responses.UpdatedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &reviewResp)
	assert.Equal(t, 1, reviewResp.Updated)

	g, err := MockController.Repo.GetGenerationOutput(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generationoutput.GalleryStatusAccepted, g.GalleryStatus)

	// ! Can reject generation
	reqBody = requests.ReviewGalleryRequest{
		Action:              requests.GalleryRejectAction,
		GenerationOutputIDs: []uuid.UUID{targetGUid},
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &reviewResp)
	assert.Equal(t, 1, reviewResp.Updated)

	g, err = MockController.Repo.GetGenerationOutput(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generationoutput.GalleryStatusRejected, g.GalleryStatus)
}

func TestHandleDeleteGeneration(t *testing.T) {
	ctx := context.Background()
	// Create mock generation
	targetG, err := MockController.Repo.CreateMockGenerationForDeletion(ctx)
	// Create generation output
	targetGOutput, err := MockController.Repo.DB.GenerationOutput.Create().SetGenerationID(targetG.ID).SetImagePath("s3://hello/world.png").SetUpscaledImagePath("s3://hello/upscaled.png").Save(ctx)
	assert.Nil(t, err)
	assert.Nil(t, targetGOutput.DeletedAt)
	// Create upscale output
	targetUpscale, err := MockController.Repo.CreateMockUpscaleForDeletion(ctx)
	targetUpscaleOutput, err := MockController.Repo.DB.UpscaleOutput.Create().SetImagePath("s3://hello/upscaled.png").SetUpscaleID(targetUpscale.ID).Save(ctx)
	assert.Nil(t, err)

	// ! Can delete generation
	reqBody := requests.DeleteGenerationRequest{
		GenerationOutputIDs: []uuid.UUID{targetGOutput.ID},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleDeleteGenerationOutput(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var deleteResp responses.DeletedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &deleteResp)
	assert.Equal(t, 1, deleteResp.Deleted)

	deletedGOutput, err := MockController.Repo.GetGenerationOutput(targetGOutput.ID)
	assert.Nil(t, err)
	assert.NotNil(t, deletedGOutput.DeletedAt)

	upscaledOutput, err := MockController.Repo.GetUpscaleOutputWithPath(targetUpscaleOutput.ImagePath)
	assert.Nil(t, err)
	assert.NotNil(t, upscaledOutput.DeletedAt)

	// Cleanup
	err = MockController.Repo.DB.GenerationOutput.DeleteOne(deletedGOutput).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.Generation.DeleteOne(targetG).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.UpscaleOutput.DeleteOne(upscaledOutput).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.Upscale.DeleteOne(targetUpscale).Exec(ctx)
	assert.Nil(t, err)
}
