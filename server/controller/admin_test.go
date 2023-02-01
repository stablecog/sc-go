package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stretchr/testify/assert"
)

func TestHandleGenerationDeleteAndApproveRejectGallery(t *testing.T) {
	// ! HTTP delete rejects approve action
	reqBody := requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionApprove,
		GenerationID: uuid.New(),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleGenerationDeleteAndApproveRejectGallery(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 405, resp.StatusCode)
	var errorResp map[string]string
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "Cannot use DELETE to approve/reject image", errorResp["error"])

	// ! HTTP post rejects delete action
	reqBody = requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionDelete,
		GenerationID: uuid.New(),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleGenerationDeleteAndApproveRejectGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 405, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "Cannot use POST to delete image", errorResp["error"])

	// ! Can approve generation
	generations, err := MockController.Repo.GetUserGenerations(uuid.MustParse(database.MOCK_ADMIN_UUID), 50, nil)
	assert.Nil(t, err)
	assert.NotEqual(t, generation.GalleryStatusAccepted, generations[1].GalleryStatus)
	targetGUid := generations[1].ID

	reqBody = requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionApprove,
		GenerationID: targetGUid,
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleGenerationDeleteAndApproveRejectGallery(w, req.WithContext(ctx))
	resp = w.Result()
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

	MockController.HandleGenerationDeleteAndApproveRejectGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	g, err = MockController.Repo.GetGeneration(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generation.GalleryStatusRejected, g.GalleryStatus)

	// ! Can delete generation
	reqBody = requests.AdminGalleryRequestBody{
		Action:       requests.AdminGalleryActionDelete,
		GenerationID: targetGUid,
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.HandleGenerationDeleteAndApproveRejectGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	g, err = MockController.Repo.GetGeneration(targetGUid)
	assert.NotNil(t, err)
	assert.True(t, ent.IsNotFound(err))
}
