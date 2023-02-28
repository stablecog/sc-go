package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stretchr/testify/assert"
)

func TestSubmitGenerationToGallery(t *testing.T) {
	// ! Generation that doesnt exist
	reqBody := requests.SubmitGalleryRequest{
		GenerationOutputIDs: []uuid.UUID{uuid.New()},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleSubmitGenerationToGallery(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var submitResp responses.SubmitGalleryResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &submitResp)

	assert.Equal(t, 0, submitResp.Submitted)

	// ! Generation that does exist
	// Retrieve generation output for user that is not submitted
	// Find goutput not approved
	goutput, err := MockController.Repo.DB.Generation.Query().Where(generation.UserIDEQ(uuid.MustParse(repository.MOCK_ADMIN_UUID))).QueryGenerationOutputs().Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusNotSubmitted)).First(MockController.Repo.Ctx)
	assert.Nil(t, err)

	reqBody = requests.SubmitGalleryRequest{
		GenerationOutputIDs: []uuid.UUID{goutput.ID},
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleSubmitGenerationToGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &submitResp)
	assert.Equal(t, 1, submitResp.Submitted)

	// Make sure updated in DB
	goutput, err = MockController.Repo.DB.GenerationOutput.Query().Where(generationoutput.IDEQ(goutput.ID)).First(MockController.Repo.Ctx)
	assert.Nil(t, err)
	assert.Equal(t, generationoutput.GalleryStatusSubmitted, goutput.GalleryStatus)

	// ! Generation that is already submitted
	reqBody = requests.SubmitGalleryRequest{
		GenerationOutputIDs: []uuid.UUID{goutput.ID},
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleSubmitGenerationToGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &submitResp)

	assert.Equal(t, 0, submitResp.Submitted)
}
