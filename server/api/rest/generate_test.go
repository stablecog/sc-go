package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
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

	MockController.HandleCreateGeneration(w, req)
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
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Unauthorized", respJson["error"])
}

func TestGenerateFailsWithInvalidStreamID(t *testing.T) {
	reqBody := requests.CreateGenerationRequest{
		StreamID: "invalid",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_stream_id", respJson["error"])
}

func TestGenerateEnforcesNumOutputsChange(t *testing.T) {
	reqBody := requests.CreateGenerationRequest{
		StreamID:      MockSSEId,
		Height:        utils.ToPtr[int32](768),
		Width:         utils.ToPtr[int32](768),
		GuidanceScale: utils.ToPtr[float32](7),
		// Minimum is not enforced since it should default to 1
		NumOutputs: utils.ToPtr[int32](-1),
		ModelId:    utils.ToPtr(uuid.MustParse("9b7d9d0e-9b7d-9d0e-9b7d-9d0e9b7d9d0e")),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_model_id", respJson["error"])

	// ! Max
	reqBody = requests.CreateGenerationRequest{
		StreamID:      MockSSEId,
		Height:        utils.ToPtr[int32](768),
		Width:         utils.ToPtr[int32](768),
		NumOutputs:    utils.ToPtr[int32](shared.MAX_GENERATE_NUM_OUTPUTS + 1),
		GuidanceScale: utils.ToPtr[float32](7),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, fmt.Sprintf("Number of outputs can't be more than %d", shared.MAX_GENERATE_NUM_OUTPUTS), respJson["error"])
}

func TestGenerateEnforcesMaxWidthMaxHeight(t *testing.T) {
	reqBody := requests.CreateGenerationRequest{
		StreamID:      MockSSEId,
		Height:        utils.ToPtr[int32](shared.MAX_GENERATE_HEIGHT + 1),
		Width:         utils.ToPtr[int32](shared.MAX_GENERATE_WIDTH),
		GuidanceScale: utils.ToPtr[float32](7),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, fmt.Sprintf("Height is too large, max is: %d", shared.MAX_GENERATE_HEIGHT), respJson["error"])

	// ! Width
	reqBody = requests.CreateGenerationRequest{
		StreamID: MockSSEId,
		Height:   utils.ToPtr[int32](shared.MAX_GENERATE_HEIGHT),
		Width:    utils.ToPtr[int32](shared.MAX_GENERATE_WIDTH + 1),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, fmt.Sprintf("Width is too large, max is: %d", shared.MAX_GENERATE_WIDTH), respJson["error"])
}

func TestGenerateValidationsSkippedForSuperAdmin(t *testing.T) {
	reqBody := requests.CreateGenerationRequest{
		StreamID:      MockSSEId,
		Height:        utils.ToPtr[int32](shared.MAX_GENERATE_HEIGHT + 1),
		Width:         utils.ToPtr[int32](shared.MAX_GENERATE_WIDTH),
		GuidanceScale: utils.ToPtr[float32](7),
		SchedulerId:   utils.ToPtr(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_model_or_scheduler", respJson["error"])

	// ! Width
	reqBody = requests.CreateGenerationRequest{
		StreamID:    MockSSEId,
		Height:      utils.ToPtr[int32](shared.MAX_GENERATE_HEIGHT),
		Width:       utils.ToPtr[int32](shared.MAX_GENERATE_WIDTH + 1),
		SchedulerId: utils.ToPtr(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_model_or_scheduler", respJson["error"])
}

func TestGenerateRejectsInvalidModelOrScheduler(t *testing.T) {
	// ! invalid_scheduler_id
	reqBody := requests.CreateGenerationRequest{
		StreamID:      MockSSEId,
		Height:        utils.ToPtr[int32](768),
		Width:         utils.ToPtr[int32](768),
		SchedulerId:   utils.ToPtr(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
		ModelId:       utils.ToPtr(uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID)),
		GuidanceScale: utils.ToPtr[float32](7),
		NumOutputs:    utils.ToPtr[int32](1),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_scheduler_id", respJson["error"])

	// ! invalid_model_id
	reqBody = requests.CreateGenerationRequest{
		StreamID:      MockSSEId,
		Height:        utils.ToPtr[int32](768),
		Width:         utils.ToPtr[int32](768),
		SchedulerId:   utils.ToPtr(uuid.MustParse(repository.MOCK_SCHEDULER_ID)),
		ModelId:       utils.ToPtr(uuid.MustParse("00000000-0000-0000-0000-000000000000")),
		GuidanceScale: utils.ToPtr[float32](7),
		NumOutputs:    utils.ToPtr[int32](1),
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_model_id", respJson["error"])
}

func TestGenerateNoCredits(t *testing.T) {
	// ! Perfectly valid request
	reqBody := requests.CreateGenerationRequest{
		StreamID:       MockSSEId,
		Height:         utils.ToPtr[int32](512),
		Width:          utils.ToPtr[int32](512),
		SchedulerId:    utils.ToPtr(uuid.MustParse(repository.MOCK_SCHEDULER_ID)),
		ModelId:        utils.ToPtr(uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID)),
		NumOutputs:     utils.ToPtr[int32](1),
		InferenceSteps: utils.ToPtr[int32](shared.MAX_GENERATE_INTERFERENCE_STEPS_FREE + 1),
		GuidanceScale:  utils.ToPtr[float32](7),
		Prompt:         "A portrait of a cat by Van Gogh",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NO_CREDITS_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var errResp map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &errResp)
	assert.Equal(t, "insufficient_credits", errResp["error"])
}

func TestGenerateValidRequest(t *testing.T) {
	// ! Perfectly valid request
	reqBody := requests.CreateGenerationRequest{
		StreamID:       MockSSEId,
		Height:         utils.ToPtr[int32](512),
		Width:          utils.ToPtr[int32](512),
		SchedulerId:    utils.ToPtr(uuid.MustParse(repository.MOCK_SCHEDULER_ID)),
		ModelId:        utils.ToPtr(uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID)),
		NumOutputs:     utils.ToPtr[int32](1),
		GuidanceScale:  utils.ToPtr[float32](7),
		InferenceSteps: utils.ToPtr[int32](shared.MAX_GENERATE_INTERFERENCE_STEPS_FREE + 1),
		Prompt:         "A portrait of a cat by Van Gogh",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleCreateGeneration(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var generateResp responses.TaskQueuedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &generateResp)

	// Make sure valid uuid
	_, err := uuid.Parse(generateResp.ID)
	assert.Nil(t, err)
}
