package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
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

	reqBody := requests.AdminGalleryRequestBody{
		Action:              requests.AdminGalleryActionApprove,
		GenerationOutputIDs: []uuid.UUID{targetGUid},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var reviewResp responses.AdminGalleryResponseBody
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &reviewResp)
	assert.Equal(t, 1, reviewResp.Updated)

	g, err := MockController.Repo.GetGenerationOutput(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generationoutput.GalleryStatusAccepted, g.GalleryStatus)

	// ! Can reject generation
	reqBody = requests.AdminGalleryRequestBody{
		Action:              requests.AdminGalleryActionReject,
		GenerationOutputIDs: []uuid.UUID{targetGUid},
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

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
	targetGUid := targetG.ID

	// ! Can delete generation
	reqBody := requests.AdminGenerationDeleteRequest{
		GenerationIDs: []uuid.UUID{targetGUid},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleDeleteGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var deleteResp responses.AdminDeleteResponseBody
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &deleteResp)
	assert.Equal(t, 1, deleteResp.Deleted)

	_, err = MockController.Repo.GetGeneration(targetGUid)
	assert.NotNil(t, err)
	assert.True(t, ent.IsNotFound(err))
}
