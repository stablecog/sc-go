package rest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stretchr/testify/assert"
)

func TestUpscaleUnauthorizedIfUserIdMissingInContext(t *testing.T) {
	reqBody := map[string]interface{}{
		"upscale": "upscale",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	MockController.HandleUpscale(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Unauthorized", respJson["error"])
}

func TestUpscaleUnauthorizedIfUserIdNotUuid(t *testing.T) {
	reqBody := map[string]interface{}{
		"upscale": "upscale",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", "not-uuid")
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Unauthorized", respJson["error"])
}

func TestUpscaleFailsWithInvalidStreamID(t *testing.T) {
	reqBody := requests.CreateUpscaleRequest{
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

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_stream_id", respJson["error"])
}

func TestUpscaleEnforcesType(t *testing.T) {
	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     "invalid",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "Invalid upscale type, should be from_image or from_output", respJson["error"])
}

func TestUpscaleErrorsBadURL(t *testing.T) {
	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     requests.UpscaleRequestTypeImage,
		Input:    "not-a-url",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_image_url", respJson["error"])
}

func TestUpscaleErrorsBadOutputID(t *testing.T) {
	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     requests.UpscaleRequestTypeOutput,
		Input:    "not-a-uuid",
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_output_id", respJson["error"])
}

func TestUpscaleRejectsInvalidModel(t *testing.T) {
	// ! invalid_model_id
	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     requests.UpscaleRequestTypeImage,
		Input:    "https://example.com/image.png",
		ModelId:  uuid.MustParse("00000000-0000-0000-0000-000000000000"),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NORMAL_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_NORMAL_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "invalid_model_id", respJson["error"])
}

func TestUpscaleFailsIfNoCredits(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "https://example.com/image.png",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", "40")
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "https://example.com/image.png",
		func(req *http.Request) (*http.Response, error) {
			const TestPNG = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEUAAACnej3aAAAAAXRSTlMAQObYZgAAAApJREFUCNdjYAAAAAIAAeIhvDMAAAAASUVORK5CYII="
			i := strings.Index(TestPNG, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestPNG[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/png")
			return resp, nil
		},
	)

	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     requests.UpscaleRequestTypeImage,
		Input:    "https://example.com/image.png",
		ModelId:  uuid.MustParse(repository.MOCK_UPSCALE_MODEL_ID),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_NO_CREDITS_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
	var respJson map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &respJson)

	assert.Equal(t, "insufficient_credits", respJson["error"])
}

func TestUpscaleFromURL(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "https://example.com/image.png",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", "40")
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "https://example.com/image.png",
		func(req *http.Request) (*http.Response, error) {
			const TestPNG = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEUAAACnej3aAAAAAXRSTlMAQObYZgAAAApJREFUCNdjYAAAAAIAAeIhvDMAAAAASUVORK5CYII="
			i := strings.Index(TestPNG, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestPNG[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/png")
			return resp, nil
		},
	)

	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     requests.UpscaleRequestTypeImage,
		Input:    "https://example.com/image.png",
		ModelId:  uuid.MustParse(repository.MOCK_UPSCALE_MODEL_ID),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var upscaleResp responses.TaskQueuedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &upscaleResp)

	// Make sure valid uuid
	_, err := uuid.Parse(upscaleResp.ID)
	assert.Nil(t, err)
}

func TestUpscaleFromOutput(t *testing.T) {
	// Get an output
	output, err := MockController.Repo.DB.Generation.Query().Where(generation.UserIDEQ(uuid.MustParse(repository.MOCK_ADMIN_UUID))).QueryGenerationOutputs().First(context.Background())
	assert.Nil(t, err)

	reqBody := requests.CreateUpscaleRequest{
		StreamID: MockSSEId,
		Type:     requests.UpscaleRequestTypeOutput,
		Input:    output.ID.String(),
		ModelId:  uuid.MustParse(repository.MOCK_UPSCALE_MODEL_ID),
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleUpscale(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var upscaleResp responses.TaskQueuedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &upscaleResp)

	// Make sure valid uuid
	_, err = uuid.Parse(upscaleResp.ID)
	assert.Nil(t, err)
}
