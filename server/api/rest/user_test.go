package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/requests"
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
	var genResponse repository.UserGenerationQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Generations, 0)
	assert.Equal(t, 0, *genResponse.Total)
	assert.Nil(t, genResponse.Next)
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
	var genResponse repository.UserGenerationQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Equal(t, 4, *genResponse.Total)
	assert.Len(t, genResponse.Generations, 4)
	assert.Nil(t, genResponse.Next)

	// They should be in order of how we mocked them (descending)
	assert.Equal(t, "This is a prompt 4", genResponse.Generations[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse.Generations[0].Status)
	assert.NotNil(t, genResponse.Generations[0].StartedAt)
	assert.Nil(t, genResponse.Generations[0].CompletedAt)
	assert.Empty(t, genResponse.Generations[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Generations[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Generations[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[0].Width)
	assert.Equal(t, int32(512), genResponse.Generations[0].Height)
	assert.Len(t, genResponse.Generations[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse.Generations[0].Seed)

	assert.Equal(t, "This is a prompt 3", genResponse.Generations[1].Prompt)
	assert.Equal(t, string(generation.StatusFailed), genResponse.Generations[1].Status)
	assert.NotNil(t, genResponse.Generations[1].StartedAt)
	assert.Nil(t, genResponse.Generations[1].CompletedAt)
	assert.Empty(t, genResponse.Generations[1].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Generations[1].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Generations[1].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[1].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[1].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[1].Width)
	assert.Equal(t, int32(512), genResponse.Generations[1].Height)
	assert.Len(t, genResponse.Generations[1].Outputs, 0)
	assert.Equal(t, 1234, genResponse.Generations[1].Seed)

	assert.Equal(t, "This is a prompt 2", genResponse.Generations[2].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Generations[2].Status)
	assert.NotNil(t, genResponse.Generations[2].StartedAt)
	assert.NotNil(t, genResponse.Generations[2].CompletedAt)
	assert.Empty(t, genResponse.Generations[2].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Generations[2].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Generations[2].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[2].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[2].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[2].Width)
	assert.Equal(t, int32(512), genResponse.Generations[2].Height)
	assert.Len(t, genResponse.Generations[2].Outputs, 3)
	assert.Equal(t, 1234, genResponse.Generations[2].Seed)

	assert.Equal(t, "This is a prompt", genResponse.Generations[3].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Generations[3].Status)
	assert.NotNil(t, genResponse.Generations[3].StartedAt)
	assert.NotNil(t, genResponse.Generations[3].CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Generations[3].NegativePrompt)
	assert.Equal(t, int32(11), genResponse.Generations[3].InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Generations[3].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[3].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[3].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[3].Width)
	assert.Equal(t, int32(512), genResponse.Generations[3].Height)
	assert.Len(t, genResponse.Generations[3].Outputs, 3)
	assert.Equal(t, 1234, genResponse.Generations[3].Seed)
}

func TestHandleQueryGenerationsCursor(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.UserGenerationQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Generations, 4)
	assert.Nil(t, genResponse.Next)

	// Get tiemstamp of first item so we can exclude it in "second page"
	assert.Equal(t, "This is a prompt 4", genResponse.Generations[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse.Generations[0].Status)
	assert.NotNil(t, genResponse.Generations[0].StartedAt)
	assert.Nil(t, genResponse.Generations[0].CompletedAt)
	assert.Empty(t, genResponse.Generations[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Generations[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Generations[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[0].Width)
	assert.Equal(t, int32(512), genResponse.Generations[0].Height)
	assert.Len(t, genResponse.Generations[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse.Generations[0].Seed)

	// With cursor off most recent item, we should get 3 items
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/gens?cursor=%s", utils.TimeToIsoString(genResponse.Generations[0].CreatedAt)), nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Generations, 3)
	assert.Equal(t, "This is a prompt 3", genResponse.Generations[0].Prompt)
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
	var genResponse repository.UserGenerationQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Generations, 1)
	assert.Equal(t, *genResponse.Next, genResponse.Generations[0].CreatedAt)

	assert.Equal(t, "This is a prompt 4", genResponse.Generations[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse.Generations[0].Status)
	assert.NotNil(t, genResponse.Generations[0].StartedAt)
	assert.Nil(t, genResponse.Generations[0].CompletedAt)
	assert.Empty(t, genResponse.Generations[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Generations[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Generations[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[0].Width)
	assert.Equal(t, int32(512), genResponse.Generations[0].Height)
	assert.Len(t, genResponse.Generations[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse.Generations[0].Seed)
}

// Test some filter params
func TestHandleQueryGenerationsFilters(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?inference_steps=11&min_guidance_scale=2", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.UserGenerationQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Generations, 1)
	assert.Nil(t, genResponse.Next)

	assert.Equal(t, "This is a prompt", genResponse.Generations[0].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Generations[0].Status)
	assert.NotNil(t, genResponse.Generations[0].StartedAt)
	assert.NotNil(t, genResponse.Generations[0].CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Generations[0].NegativePrompt)
	assert.Equal(t, int32(11), genResponse.Generations[0].InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Generations[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Generations[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Generations[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Generations[0].Width)
	assert.Equal(t, int32(512), genResponse.Generations[0].Height)
	assert.Len(t, genResponse.Generations[0].Outputs, 3)
	assert.Equal(t, 1234, genResponse.Generations[0].Seed)
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

func TestHandleQueryGenerationsBadCursor(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?cursor=HelloWorld", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var errorResp map[string]string
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "cursor must be a valid iso time string", errorResp["error"])
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

// Parse query filters from url params
func TestParseQueryGenerationFilters(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&succeeded_only=true"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	filters, err := ParseQueryGenerationFilters(values)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), filters.MinWidth)
	assert.Equal(t, int32(5), filters.MaxWidth)
	assert.Equal(t, int32(6), filters.MinHeight)
	assert.Equal(t, int32(7), filters.MaxHeight)
	assert.Equal(t, int32(2), filters.MinInferenceSteps)
	assert.Equal(t, int32(3), filters.MaxInferenceSteps)
	assert.Equal(t, float32(2), filters.MinGuidanceScale)
	assert.Equal(t, float32(4), filters.MaxGuidanceScale)
	assert.Equal(t, []int32{512, 768}, filters.Widths)
	assert.Equal(t, []int32{512}, filters.Heights)
	assert.Equal(t, []int32{30}, filters.InferenceSteps)
	assert.Equal(t, []float32{5}, filters.GuidanceScales)
	assert.Equal(t, []uuid.UUID{uuid.MustParse("e07ad712-41ad-4ff7-8727-faf0d91e4c4e"), uuid.MustParse("c09aaf4d-2d78-4281-89aa-88d5d0a5d70b")}, filters.SchedulerIDs)
	assert.Equal(t, []uuid.UUID{uuid.MustParse("49d75ae2-5407-40d9-8c02-0c44ba08f358")}, filters.ModelIDs)
	assert.Equal(t, true, filters.SucceededOnly)
	// Default descending
	assert.Equal(t, requests.UserGenerationQueryOrderDescending, filters.Order)
}

func TestParseQueryGenerationFilterError(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&succeeded_only=true&order=invalid"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	_, err = ParseQueryGenerationFilters(values)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid order: invalid expected 'asc' or 'desc'", err.Error())
}
