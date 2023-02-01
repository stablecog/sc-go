package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/server/requests"
	"github.com/stablecog/go-apps/server/responses"
	"github.com/stablecog/go-apps/shared"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUnauthorizedIfUserIdMissingInContext(t *testing.T) {
	reqBody := map[string]interface{}{
		"generate": "generate",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	MockController.PostGenerate(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Unauthorized", respJson["error"])
}

func TestGenerateUnauthorizedIfUserIdNotUuid(t *testing.T) {
	reqBody := map[string]interface{}{
		"generate": "generate",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", "not-uuid")

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Unauthorized", respJson["error"])
}

func TestGenerateFailsWithInvalidWebsocketID(t *testing.T) {
	reqBody := requests.GenerateRequestBody{
		WebsocketId: "invalid",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Invalid websocket ID", respJson["error"])
}

func TestGenerateEnforcesMaxWidthMaxHeight(t *testing.T) {
	reqBody := requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT + 1,
		Width:       shared.MAX_GENERATE_WIDTH,
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, fmt.Sprintf("Height is too large, max is: %d", shared.MAX_GENERATE_HEIGHT), respJson["error"])

	// ! Width
	reqBody = requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT,
		Width:       shared.MAX_GENERATE_WIDTH + 1,
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, fmt.Sprintf("Width is too large, max is: %d", shared.MAX_GENERATE_WIDTH), respJson["error"])
}

func TestGenerateRejectsInvalidModelOrScheduler(t *testing.T) {
	// ! Invalid scheduler ID
	reqBody := requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT,
		Width:       shared.MAX_GENERATE_WIDTH,
		SchedulerId: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		ModelId:     uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_FREE),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Invalid scheduler ID", respJson["error"])

	// ! Invalid model ID
	reqBody = requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT,
		Width:       shared.MAX_GENERATE_WIDTH,
		SchedulerId: uuid.MustParse(database.MOCK_SCHEDULER_ID_FREE),
		ModelId:     uuid.MustParse("00000000-0000-0000-0000-000000000000"),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Invalid model ID", respJson["error"])
}

// Test all of the restrictons for free users
func TestGenerateProRestrictions(t *testing.T) {
	// ! PRO only model
	reqBody := requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT,
		Width:       shared.MAX_GENERATE_WIDTH,
		SchedulerId: uuid.MustParse(database.MOCK_SCHEDULER_ID_FREE),
		ModelId:     uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_PRO),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_FREE_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "That model is not available on the free plan :(", respJson["error"])

	// ! PRO only scheduler
	reqBody = requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT,
		Width:       shared.MAX_GENERATE_WIDTH,
		SchedulerId: uuid.MustParse(database.MOCK_SCHEDULER_ID_PRO),
		ModelId:     uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_FREE),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_FREE_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "That scheduler is not available on the free plan :(", respJson["error"])

	// ! PRO only height
	reqBody = requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT_FREE + 1,
		Width:       shared.MAX_GENERATE_WIDTH,
		SchedulerId: uuid.MustParse(database.MOCK_SCHEDULER_ID_FREE),
		ModelId:     uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_FREE),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_FREE_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "That generation height is not available on the free plan :(", respJson["error"])

	// ! PRO only width
	reqBody = requests.GenerateRequestBody{
		WebsocketId: MockWSId,
		Height:      shared.MAX_GENERATE_HEIGHT_FREE,
		Width:       shared.MAX_GENERATE_WIDTH_FREE + 1,
		SchedulerId: uuid.MustParse(database.MOCK_SCHEDULER_ID_FREE),
		ModelId:     uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_FREE),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_FREE_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "That generation width is not available on the free plan :(", respJson["error"])

	// ! PRO only interference steps
	reqBody = requests.GenerateRequestBody{
		WebsocketId:       MockWSId,
		Height:            shared.MAX_GENERATE_HEIGHT_FREE,
		Width:             shared.MAX_GENERATE_WIDTH_FREE,
		NumInferenceSteps: shared.MAX_GENERATE_INTERFERENCE_STEPS_FREE + 1,
		SchedulerId:       uuid.MustParse(database.MOCK_SCHEDULER_ID_FREE),
		ModelId:           uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_FREE),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_FREE_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "That number of inference steps is not available on the free plan :(", respJson["error"])
}

func TestGenerateValidRequest(t *testing.T) {
	// ! Perfectly valid request
	reqBody := requests.GenerateRequestBody{
		WebsocketId:       MockWSId,
		Height:            shared.MAX_GENERATE_HEIGHT,
		Width:             shared.MAX_GENERATE_WIDTH,
		SchedulerId:       uuid.MustParse(database.MOCK_SCHEDULER_ID_FREE),
		ModelId:           uuid.MustParse(database.MOCK_GENERATION_MODEL_ID_PRO),
		NumInferenceSteps: shared.MAX_GENERATE_INTERFERENCE_STEPS_FREE + 1,
		Prompt:            "A portrait of a cat by Van Gogh",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_PRO_UUID)

	MockController.PostGenerate(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var generateResp responses.GenerateResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &generateResp)

	// Make sure valid uuid
	_, err := uuid.Parse(generateResp.ID)
	assert.Nil(t, err)

	// make sure we have this ID on our map
	assert.Equal(t, MockWSId, MockController.CogRequestWebsocketConnMap.Get(generateResp.ID))
}

func TestSubmitGenerationToGallery(t *testing.T) {
	// ! Generation that doesnt exist
	reqBody := requests.GenerateSubmitToGalleryRequestBody{
		GenerationID: uuid.New(),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", database.MOCK_PRO_UUID)

	MockController.PostSubmitGenerationToGallery(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var errorResp map[string]string
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "Generation not found", errorResp["error"])

	// ! Generation that does exist
	// Retrieve generation
	generations, err := MockController.Repo.GetUserGenerations(uuid.MustParse(database.MOCK_ADMIN_UUID), 50, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, generations)
	assert.Equal(t, generation.GalleryStatusNotSubmitted, generations[0].GalleryStatus)

	reqBody = requests.GenerateSubmitToGalleryRequestBody{
		GenerationID: generations[0].ID,
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostSubmitGenerationToGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	// Make sure updated in DB
	generations, err = MockController.Repo.GetUserGenerations(uuid.MustParse(database.MOCK_ADMIN_UUID), 50, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, generations)
	assert.Equal(t, generation.GalleryStatusSubmitted, generations[0].GalleryStatus)

	// ! Generation that is already submitted
	// Retrieve generation
	generations, err = MockController.Repo.GetUserGenerations(uuid.MustParse(database.MOCK_ADMIN_UUID), 50, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, generations)
	assert.Equal(t, generation.GalleryStatusSubmitted, generations[0].GalleryStatus)

	reqBody = requests.GenerateSubmitToGalleryRequestBody{
		GenerationID: generations[0].ID,
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", database.MOCK_ADMIN_UUID)

	MockController.PostSubmitGenerationToGallery(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errorResp)

	assert.Equal(t, "Generation already submitted to gallery", errorResp["error"])
}
