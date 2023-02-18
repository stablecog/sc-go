package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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
	var genResponse repository.GenerationQueryWithOutputsMeta
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
	var genResponse repository.GenerationQueryWithOutputsMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Equal(t, 6, *genResponse.Total)
	assert.Len(t, genResponse.Outputs, 6)
	assert.Nil(t, genResponse.Next)

	// They should be in order of how we mocked them (descending)
	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Empty(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)
	assert.Len(t, genResponse.Outputs[0].Generation.Outputs, 3)
	assert.Equal(t, "output_6", genResponse.Outputs[0].Generation.Outputs[0].ImageUrl)
	assert.Equal(t, "output_5", genResponse.Outputs[0].Generation.Outputs[1].ImageUrl)
	assert.Equal(t, "output_4", genResponse.Outputs[0].Generation.Outputs[2].ImageUrl)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[1].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[1].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[1].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[1].Generation.CompletedAt)
	assert.Empty(t, genResponse.Outputs[1].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[1].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[1].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[1].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[1].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[1].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[1].Generation.Height)
	assert.Equal(t, "output_5", genResponse.Outputs[1].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[1].Generation.Seed)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[2].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[2].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[2].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[2].Generation.CompletedAt)
	assert.Empty(t, genResponse.Outputs[2].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[2].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[2].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[2].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[2].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[2].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[2].Generation.Height)
	assert.Equal(t, "output_4", genResponse.Outputs[2].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[2].Generation.Seed)

	assert.Equal(t, "This is a prompt", genResponse.Outputs[3].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[3].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[3].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[3].Generation.CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Outputs[3].Generation.NegativePrompt)
	assert.Equal(t, int32(11), genResponse.Outputs[3].Generation.InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Outputs[3].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[3].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[3].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[3].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[3].Generation.Height)
	assert.Equal(t, "output_3", genResponse.Outputs[3].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[3].Generation.Seed)
}

func TestHandleQueryGenerationsCursor(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?upscaled=not", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.GenerationQueryWithOutputsMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Outputs, 6)
	assert.Nil(t, genResponse.Next)

	// Get tiemstamp of first item so we can exclude it in "second page"
	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Empty(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)

	// With cursor off most recent item, we should get next items
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/gens?cursor=%s", utils.TimeToIsoString(genResponse.Outputs[0].Generation.CreatedAt)), nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	MockController.HandleQueryGenerations(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	genResponse = repository.GenerationQueryWithOutputsMeta{}
	json.Unmarshal(respBody, &genResponse)

	assert.Nil(t, genResponse.Total)
	assert.Len(t, genResponse.Outputs, 3)
	assert.Equal(t, "This is a prompt", genResponse.Outputs[0].Generation.Prompt)
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
	var genResponse repository.GenerationQueryWithOutputsMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Outputs, 1)
	assert.Equal(t, *genResponse.Next, genResponse.Outputs[0].Generation.CreatedAt)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Empty(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)
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
	var genResponse repository.GenerationQueryWithOutputsMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Outputs, 3)
	assert.Nil(t, genResponse.Next)

	assert.Equal(t, "This is a prompt", genResponse.Outputs[0].Generation.Prompt)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(11), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "output_3", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)
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
	var creditResp responses.QueryCreditsResponse
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
	var creditResp responses.QueryCreditsResponse
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
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&upscaled=only&start_dt=2021-01-01T00:00:00Z"
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
	assert.Equal(t, requests.UpscaleStatusOnly, filters.UpscaleStatus)
	assert.NotNil(t, filters.StartDt)
	assert.Equal(t, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), *filters.StartDt)
	// Default descending
	assert.Equal(t, requests.SortOrderDescending, filters.Order)
}

func TestParseQueryGenerationFilterError(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&order=invalid"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	_, err = ParseQueryGenerationFilters(values)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid order: 'invalid' expected 'asc' or 'desc'", err.Error())
}

func TestHandleDeleteGenerationForUser(t *testing.T) {
	ctx := context.Background()
	// Create mock generation
	targetG, err := MockController.Repo.CreateMockGenerationForDeletion(ctx)
	// Create generation output
	targetGOutput, err := MockController.Repo.DB.GenerationOutput.Create().SetGenerationID(targetG.ID).SetImagePath("s3://hello/world.png").Save(ctx)
	assert.Nil(t, err)
	assert.Nil(t, targetGOutput.DeletedAt)

	// ! Can not delete generation unless it belongs to user
	reqBody := requests.DeleteGenerationRequest{
		GenerationOutputIDs: []uuid.UUID{targetGOutput.ID},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)

	MockController.HandleDeleteGenerationOutputForUser(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var deleteResp responses.DeletedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &deleteResp)
	assert.Equal(t, 0, deleteResp.Deleted)

	deletedGOutput, err := MockController.Repo.GetGenerationOutput(targetGOutput.ID)
	assert.Nil(t, err)
	assert.Nil(t, deletedGOutput.DeletedAt)

	// ! Can delete generation if it belongs to user
	// Build request
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)

	MockController.HandleDeleteGenerationOutputForUser(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &deleteResp)
	assert.Equal(t, 1, deleteResp.Deleted)

	deletedGOutput, err = MockController.Repo.GetGenerationOutput(targetGOutput.ID)
	assert.Nil(t, err)
	assert.NotNil(t, deletedGOutput.DeletedAt)

	// Cleanup
	err = MockController.Repo.DB.GenerationOutput.DeleteOne(deletedGOutput).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.Generation.DeleteOne(targetG).Exec(ctx)
	assert.Nil(t, err)
}
