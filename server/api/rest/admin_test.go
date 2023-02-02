package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stretchr/testify/assert"
)

func TestHandleReviewGallerySubmission(t *testing.T) {

	// ! Can approve generation
	// Retrieve generations
	generations, err := MockController.Repo.GetUserGenerations(uuid.MustParse(database.MOCK_ADMIN_UUID), 50, nil)
	assert.Nil(t, err)
	targetGUid := generations[1].ID

	reqBody := requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionApprove,
		GenerationID: targetGUid,
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	g, err := MockController.Repo.GetGeneration(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generation.GalleryStatusAccepted, g.GalleryStatus)

	// ! Can reject generation
	reqBody = requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionReject,
		GenerationID: targetGUid,
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	g, err = MockController.Repo.GetGeneration(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generation.GalleryStatusRejected, g.GalleryStatus)
}

func TestHandleDeleteGeneration(t *testing.T) {
	ctx := context.Background()
	// Create mock generation
	targetG, err := database.CreateMockGenerationForDeletion(ctx, MockController.Repo)
	targetGUid := targetG.ID

	// ! Can delete generation
	reqBody := requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionDelete,
		GenerationID: targetGUid,
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleDeleteGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	_, err = MockController.Repo.GetGeneration(targetGUid)
	assert.NotNil(t, err)
	assert.True(t, ent.IsNotFound(err))
}
