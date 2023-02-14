package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

// Retrieving for user that has no generations
func TestHandleQueryGenerationsDontExist(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NO_CREDITS_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse []repository.UserGenerationQueryResult
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse, 0)
}

func TestHandleQueryGenerationsDefaultParams(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse []repository.UserGenerationQueryResult
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse, 4)

	// They should be in order of how we mocked them (descending)
	assert.Equal(t, "This is a prompt 4", genResponse[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse[0].Status)
	assert.NotNil(t, genResponse[0].StartedAt)
	assert.Nil(t, genResponse[0].CompletedAt)
	assert.Empty(t, genResponse[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse[0].Width)
	assert.Equal(t, int32(512), genResponse[0].Height)
	assert.Len(t, genResponse[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse[0].Seed)

	assert.Equal(t, "This is a prompt 3", genResponse[1].Prompt)
	assert.Equal(t, string(generation.StatusFailed), genResponse[1].Status)
	assert.NotNil(t, genResponse[1].StartedAt)
	assert.Nil(t, genResponse[1].CompletedAt)
	assert.Empty(t, genResponse[1].NegativePrompt)
	assert.Equal(t, int32(30), genResponse[1].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse[1].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse[1].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse[1].SchedulerID)
	assert.Equal(t, int32(512), genResponse[1].Width)
	assert.Equal(t, int32(512), genResponse[1].Height)
	assert.Len(t, genResponse[1].Outputs, 0)
	assert.Equal(t, 1234, genResponse[1].Seed)

	assert.Equal(t, "This is a prompt 2", genResponse[2].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse[2].Status)
	assert.NotNil(t, genResponse[2].StartedAt)
	assert.NotNil(t, genResponse[2].CompletedAt)
	assert.Empty(t, genResponse[2].NegativePrompt)
	assert.Equal(t, int32(30), genResponse[2].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse[2].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse[2].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse[2].SchedulerID)
	assert.Equal(t, int32(512), genResponse[2].Width)
	assert.Equal(t, int32(512), genResponse[2].Height)
	assert.Len(t, genResponse[2].Outputs, 3)
	assert.Equal(t, 1234, genResponse[2].Seed)

	assert.Equal(t, "This is a prompt", genResponse[3].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse[3].Status)
	assert.NotNil(t, genResponse[3].StartedAt)
	assert.NotNil(t, genResponse[3].CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse[3].NegativePrompt)
	assert.Equal(t, int32(30), genResponse[3].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse[3].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse[3].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse[3].SchedulerID)
	assert.Equal(t, int32(512), genResponse[3].Width)
	assert.Equal(t, int32(512), genResponse[3].Height)
	assert.Len(t, genResponse[3].Outputs, 3)
	assert.Equal(t, 1234, genResponse[3].Seed)
}

func TestHandleQueryGenerationsOffset(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse []repository.UserGenerationQueryResult
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse, 4)

	// Get tiemstamp of first item so we can exclude it in "second page"
	assert.Equal(t, "This is a prompt 4", genResponse[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse[0].Status)
	assert.NotNil(t, genResponse[0].StartedAt)
	assert.Nil(t, genResponse[0].CompletedAt)
	assert.Empty(t, genResponse[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse[0].Width)
	assert.Equal(t, int32(512), genResponse[0].Height)
	assert.Len(t, genResponse[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse[0].Seed)

	// With offset off most recent item, we should get 3 items
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/gens?offset=%s", utils.TimeToIsoString(genResponse[0].CreatedAt)), nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse, 3)
	assert.Equal(t, "This is a prompt 3", genResponse[0].Prompt)
}

// Test per page param
func TestHandleQueryGenerationsPerPage(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?per_page=1", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse []repository.UserGenerationQueryResult
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse, 1)

	assert.Equal(t, "This is a prompt 4", genResponse[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse[0].Status)
	assert.NotNil(t, genResponse[0].StartedAt)
	assert.Nil(t, genResponse[0].CompletedAt)
	assert.Empty(t, genResponse[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse[0].Width)
	assert.Equal(t, int32(512), genResponse[0].Height)
	assert.Len(t, genResponse[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse[0].Seed)
}

// ! Error conditions in API
func TestHandleQueryGenerationsUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?per_page=1", nil)
	req.Header.Set("Content-Type", "application/json")

	MockController.HandleQueryGenerations(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
	var genResponse map[string]string
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Equal(t, "Unauthorized", genResponse["error"])
}

func TestHandleQueryGenerationsBadPerPage(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?per_page=HelloWorld", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var errorResp map[string]string
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "per_page must be an integer", errorResp["error"])

	// Test range
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("GET", "/gens?per_page=-1", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "per_page must be between 1 and 100", errorResp["error"])

	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("GET", "/gens?per_page=101", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "per_page must be between 1 and 100", errorResp["error"])
}

func TestHandleQueryGenerationsBadOffset(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?offset=HelloWorld", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var errorResp map[string]string
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "offset must be a valid iso time string", errorResp["error"])
}

// Credits

func TestHandleQueryCreditsEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NO_CREDITS_UUID)

	MockController.HandleQueryCredits(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var creditResp responses.UserCreditsResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &creditResp)

	assert.Equal(t, int32(0), creditResp.TotalRemainingCredits)
	assert.Len(t, creditResp.Credits, 0)
}

func TestHandleQueryCredits(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ALT_UUID)

	MockController.HandleQueryCredits(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var creditResp responses.UserCreditsResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &creditResp)

	assert.Equal(t, int32(1334), creditResp.TotalRemainingCredits)
	assert.Len(t, creditResp.Credits, 2)
	assert.Equal(t, int32(100), creditResp.Credits[0].RemainingAmount)
	assert.Equal(t, "mock", creditResp.Credits[0].Type.Name)
	assert.Equal(t, int32(1234), creditResp.Credits[1].RemainingAmount)
	assert.Equal(t, "mock", creditResp.Credits[1].Type.Name)
}
