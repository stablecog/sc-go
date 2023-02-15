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

	assert.Len(t, genResponse.Outputs, 0)
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

	assert.Equal(t, 8, *genResponse.Total)
	assert.Len(t, genResponse.Outputs, 8)
	assert.Nil(t, genResponse.Next)

	// They should be in order of how we mocked them (descending)
	assert.Equal(t, "This is a prompt 4", genResponse.Outputs[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse.Outputs[0].Status)
	assert.NotNil(t, genResponse.Outputs[0].StartedAt)
	assert.Nil(t, genResponse.Outputs[0].CompletedAt)
	assert.Empty(t, genResponse.Outputs[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Height)
	assert.Equal(t, "", genResponse.Outputs[0].ImageUrl)
	assert.Nil(t, genResponse.Outputs[0].OutputID)
	assert.Equal(t, 1234, genResponse.Outputs[0].Seed)

	assert.Equal(t, "This is a prompt 3", genResponse.Outputs[1].Prompt)
	assert.Equal(t, string(generation.StatusFailed), genResponse.Outputs[1].Status)
	assert.NotNil(t, genResponse.Outputs[1].StartedAt)
	assert.Nil(t, genResponse.Outputs[1].CompletedAt)
	assert.Empty(t, genResponse.Outputs[1].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[1].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[1].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[1].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[1].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[1].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[1].Height)
	assert.Equal(t, "", genResponse.Outputs[0].ImageUrl)
	assert.Nil(t, genResponse.Outputs[0].OutputID)
	assert.Equal(t, 1234, genResponse.Outputs[1].Seed)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[2].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[2].Status)
	assert.NotNil(t, genResponse.Outputs[2].StartedAt)
	assert.NotNil(t, genResponse.Outputs[2].CompletedAt)
	assert.Empty(t, genResponse.Outputs[2].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[2].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[2].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[2].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[2].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[2].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[2].Height)
	assert.Equal(t, "output_6", genResponse.Outputs[2].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[2].Seed)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[3].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[3].Status)
	assert.NotNil(t, genResponse.Outputs[3].StartedAt)
	assert.NotNil(t, genResponse.Outputs[3].CompletedAt)
	assert.Empty(t, genResponse.Outputs[3].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[3].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[3].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[3].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[3].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[3].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[3].Height)
	assert.Equal(t, "output_5", genResponse.Outputs[3].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[3].Seed)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[4].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[4].Status)
	assert.NotNil(t, genResponse.Outputs[4].StartedAt)
	assert.NotNil(t, genResponse.Outputs[4].CompletedAt)
	assert.Empty(t, genResponse.Outputs[4].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[4].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[4].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[4].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[4].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[4].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[4].Height)
	assert.Equal(t, "output_4", genResponse.Outputs[4].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[4].Seed)

	assert.Equal(t, "This is a prompt", genResponse.Outputs[5].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[5].Status)
	assert.NotNil(t, genResponse.Outputs[5].StartedAt)
	assert.NotNil(t, genResponse.Outputs[5].CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Outputs[5].NegativePrompt)
	assert.Equal(t, int32(11), genResponse.Outputs[5].InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Outputs[5].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[5].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[5].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[5].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[5].Height)
	assert.Equal(t, "output_3", genResponse.Outputs[5].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[5].Seed)
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

	assert.Len(t, genResponse.Outputs, 8)
	assert.Nil(t, genResponse.Next)

	// Get tiemstamp of first item so we can exclude it in "second page"
	assert.Equal(t, "This is a prompt 4", genResponse.Outputs[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse.Outputs[0].Status)
	assert.NotNil(t, genResponse.Outputs[0].StartedAt)
	assert.Nil(t, genResponse.Outputs[0].CompletedAt)
	assert.Empty(t, genResponse.Outputs[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Height)
	assert.Equal(t, 1234, genResponse.Outputs[0].Seed)

	// With cursor off most recent item, we should get 3 items
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/gens?cursor=%s", utils.TimeToIsoString(genResponse.Outputs[0].CreatedAt)), nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	genResponse = repository.UserGenerationQueryMeta{}
	json.Unmarshal(respBody, &genResponse)

	assert.Nil(t, genResponse.Total)
	assert.Len(t, genResponse.Outputs, 7)
	assert.Equal(t, "This is a prompt 3", genResponse.Outputs[0].Prompt)
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

	assert.Len(t, genResponse.Outputs, 1)
	assert.Equal(t, *genResponse.Next, genResponse.Outputs[0].CreatedAt)

	assert.Equal(t, "This is a prompt 4", genResponse.Outputs[0].Prompt)
	assert.Equal(t, string(generation.StatusStarted), genResponse.Outputs[0].Status)
	assert.NotNil(t, genResponse.Outputs[0].StartedAt)
	assert.Nil(t, genResponse.Outputs[0].CompletedAt)
	assert.Empty(t, genResponse.Outputs[0].NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Height)
	// assert.Len(t, genResponse.Outputs[0].Outputs, 0)
	assert.Equal(t, 1234, genResponse.Outputs[0].Seed)
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

	assert.Len(t, genResponse.Outputs, 3)
	assert.Nil(t, genResponse.Next)

	assert.Equal(t, "This is a prompt", genResponse.Outputs[0].Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Status)
	assert.NotNil(t, genResponse.Outputs[0].StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Outputs[0].NegativePrompt)
	assert.Equal(t, int32(11), genResponse.Outputs[0].InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Outputs[0].GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Height)
	assert.Equal(t, "output_3", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Seed)
	assert.Equal(t, "output_2", genResponse.Outputs[1].ImageUrl)
	assert.Equal(t, "output_1", genResponse.Outputs[2].ImageUrl)

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
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&succeeded_only=true&upscaled=only"
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
	assert.Equal(t, requests.UserGenerationQueryUpscaleStatusOnly, filters.UpscaleStatus)
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
	assert.Equal(t, "invalid order: 'invalid' expected 'asc' or 'desc'", err.Error())
}
