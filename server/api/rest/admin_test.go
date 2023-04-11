package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestHandleReviewGallerySubmission(t *testing.T) {
	// ! Can approve generation
	var targetGUid uuid.UUID
	// Find goutput not approved
	goutput, err := MockController.Repo.DB.GenerationOutput.Query().Where(generationoutput.GalleryStatusNEQ(generationoutput.GalleryStatusApproved)).First(MockController.Repo.Ctx)
	assert.Nil(t, err)
	targetGUid = goutput.ID

	reqBody := requests.ReviewGalleryRequest{
		Action:              requests.GalleryApproveAction,
		GenerationOutputIDs: []uuid.UUID{targetGUid},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var reviewResp responses.UpdatedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &reviewResp)
	assert.Equal(t, 1, reviewResp.Updated)

	g, err := MockController.Repo.GetGenerationOutput(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generationoutput.GalleryStatusApproved, g.GalleryStatus)

	// ! Can reject generation
	reqBody = requests.ReviewGalleryRequest{
		Action:              requests.GalleryRejectAction,
		GenerationOutputIDs: []uuid.UUID{targetGUid},
	}
	body, _ = json.Marshal(reqBody)
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleReviewGallerySubmission(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &reviewResp)
	assert.Equal(t, 1, reviewResp.Updated)

	g, err = MockController.Repo.GetGenerationOutput(targetGUid)
	assert.Nil(t, err)
	assert.Equal(t, generationoutput.GalleryStatusRejected, g.GalleryStatus)
}

func TestHandleDeleteGeneration(t *testing.T) {
	ctx := context.Background()
	// Create mock generation
	targetG, err := MockController.Repo.CreateMockGenerationForDeletion(ctx)
	// Create generation output
	targetGOutput, err := MockController.Repo.DB.GenerationOutput.Create().SetGenerationID(targetG.ID).SetImagePath("s3://hello/world.png").SetUpscaledImagePath("s3://hello/upscaled.png").Save(ctx)
	assert.Nil(t, err)
	assert.Nil(t, targetGOutput.DeletedAt)
	// Create upscale output
	targetUpscale, err := MockController.Repo.CreateMockUpscaleForDeletion(ctx)
	targetUpscaleOutput, err := MockController.Repo.DB.UpscaleOutput.Create().SetImagePath("s3://hello/upscaled.png").SetUpscaleID(targetUpscale.ID).Save(ctx)
	assert.Nil(t, err)

	// ! Can NOT delete generation if gallery admiin
	reqBody := requests.DeleteGenerationRequest{
		GenerationOutputIDs: []uuid.UUID{targetGOutput.ID},
	}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameGALLERY_ADMIN.String())

	MockController.HandleDeleteGenerationOutput(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)

	// ! Can delete generation if super
	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("DELETE", "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Setup context
	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())

	MockController.HandleDeleteGenerationOutput(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var deleteResp responses.DeletedResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &deleteResp)
	assert.Equal(t, 1, deleteResp.Deleted)

	deletedGOutput, err := MockController.Repo.GetGenerationOutput(targetGOutput.ID)
	assert.Nil(t, err)
	assert.NotNil(t, deletedGOutput.DeletedAt)

	upscaledOutput, err := MockController.Repo.GetUpscaleOutputWithPath(targetUpscaleOutput.ImagePath)
	assert.Nil(t, err)
	assert.NotNil(t, upscaledOutput.DeletedAt)

	// Cleanup
	err = MockController.Repo.DB.GenerationOutput.DeleteOne(deletedGOutput).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.Generation.DeleteOne(targetG).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.UpscaleOutput.DeleteOne(upscaledOutput).Exec(ctx)
	assert.Nil(t, err)
	err = MockController.Repo.DB.Upscale.DeleteOne(targetUpscale).Exec(ctx)
	assert.Nil(t, err)
}

func TestHandleQueryGenerationsForAdminDefaultParams(t *testing.T) {
	os.Setenv("BUCKET_BASE_URL", "http://test.com/")
	defer os.Unsetenv("BUCKET_BASE_URL")

	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())

	MockController.HandleQueryGenerationsForAdmin(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.GenerationQueryWithOutputsMeta[*time.Time]
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Equal(t, 6, *genResponse.Total)
	assert.Len(t, genResponse.Outputs, 6)
	assert.Nil(t, genResponse.Next)

	// They should be in order of how we mocked them (descending)
	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Nil(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "http://test.com/output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)
	assert.Len(t, genResponse.Outputs[0].Generation.Outputs, 3)
	assert.Equal(t, "http://test.com/output_6", genResponse.Outputs[0].Generation.Outputs[0].ImageUrl)
	assert.Equal(t, "http://test.com/output_5", genResponse.Outputs[0].Generation.Outputs[1].ImageUrl)
	assert.Equal(t, "http://test.com/output_4", genResponse.Outputs[0].Generation.Outputs[2].ImageUrl)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[1].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[1].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[1].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[1].Generation.CompletedAt)
	assert.Nil(t, genResponse.Outputs[1].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[1].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[1].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[1].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[1].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[1].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[1].Generation.Height)
	assert.Equal(t, "http://test.com/output_5", genResponse.Outputs[1].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[1].Generation.Seed)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[2].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[2].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[2].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[2].Generation.CompletedAt)
	assert.Nil(t, genResponse.Outputs[2].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[2].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[2].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[2].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[2].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[2].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[2].Generation.Height)
	assert.Equal(t, "http://test.com/output_4", genResponse.Outputs[2].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[2].Generation.Seed)

	assert.Equal(t, "This is a prompt", genResponse.Outputs[3].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[3].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[3].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[3].Generation.CompletedAt)
	assert.Equal(t, "This is a negative prompt", genResponse.Outputs[3].Generation.NegativePrompt.Text)
	assert.Equal(t, int32(11), genResponse.Outputs[3].Generation.InferenceSteps)
	assert.Equal(t, float32(2.0), genResponse.Outputs[3].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[3].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[3].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[3].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[3].Generation.Height)
	assert.Equal(t, "http://test.com/output_3", genResponse.Outputs[3].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[3].Generation.Seed)
}

func TestHandleQueryGenerationsAdminCursor(t *testing.T) {
	os.Setenv("BUCKET_BASE_URL", "http://test.com/")
	defer os.Unsetenv("BUCKET_BASE_URL")

	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?upscaled=not", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())

	MockController.HandleQueryGenerationsForAdmin(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.GenerationQueryWithOutputsMeta[*time.Time]
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Outputs, 6)
	assert.Nil(t, genResponse.Next)

	// Get tiemstamp of first item so we can exclude it in "second page"
	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Nil(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "http://test.com/output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)

	// With cursor off most recent item, we should get next items
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/gens?cursor=%s", utils.TimeToIsoString(genResponse.Outputs[0].Generation.CreatedAt)), nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())

	MockController.HandleQueryGenerationsForAdmin(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	genResponse = repository.GenerationQueryWithOutputsMeta[*time.Time]{}
	json.Unmarshal(respBody, &genResponse)

	assert.Nil(t, genResponse.Total)
	assert.Len(t, genResponse.Outputs, 3)
	assert.Equal(t, "This is a prompt", genResponse.Outputs[0].Generation.Prompt.Text)
}

// Test per page param
func TestHandleQueryGenerationsAdminPerPage(t *testing.T) {
	os.Setenv("BUCKET_BASE_URL", "http://test.com/")
	defer os.Unsetenv("BUCKET_BASE_URL")
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?per_page=1", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameSUPER_ADMIN.String())

	MockController.HandleQueryGenerationsForAdmin(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.GenerationQueryWithOutputsMeta[*time.Time]
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Outputs, 1)
	assert.Equal(t, *genResponse.Next, *genResponse.Outputs[0].CreatedAt)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Nil(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "http://test.com/output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)
}

// Gallery admin can query
func TestHandleQueryGenerationsAdminGalleryLevel(t *testing.T) {
	os.Setenv("BUCKET_BASE_URL", "http://test.com/")
	defer os.Unsetenv("BUCKET_BASE_URL")
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?per_page=1", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameGALLERY_ADMIN.String())

	MockController.HandleQueryGenerationsForAdmin(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var genResponse repository.GenerationQueryWithOutputsMeta[*time.Time]
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &genResponse)

	assert.Len(t, genResponse.Outputs, 1)
	assert.Equal(t, *genResponse.Next, *genResponse.Outputs[0].CreatedAt)

	assert.Equal(t, "This is a prompt 2", genResponse.Outputs[0].Generation.Prompt.Text)
	assert.Equal(t, string(generation.StatusSucceeded), genResponse.Outputs[0].Generation.Status)
	assert.NotNil(t, genResponse.Outputs[0].Generation.StartedAt)
	assert.NotNil(t, genResponse.Outputs[0].Generation.CompletedAt)
	assert.Nil(t, genResponse.Outputs[0].Generation.NegativePrompt)
	assert.Equal(t, int32(30), genResponse.Outputs[0].Generation.InferenceSteps)
	assert.Equal(t, float32(1.0), genResponse.Outputs[0].Generation.GuidanceScale)
	assert.Equal(t, uuid.MustParse(repository.MOCK_GENERATION_MODEL_ID), genResponse.Outputs[0].Generation.ModelID)
	assert.Equal(t, uuid.MustParse(repository.MOCK_SCHEDULER_ID), genResponse.Outputs[0].Generation.SchedulerID)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Width)
	assert.Equal(t, int32(512), genResponse.Outputs[0].Generation.Height)
	assert.Equal(t, "http://test.com/output_6", genResponse.Outputs[0].ImageUrl)
	assert.Equal(t, 1234, genResponse.Outputs[0].Generation.Seed)
}

// Gallery admin cannot query private
func TestHandleQueryGenerationsAdminGalleryLevelCannotGetPrivate(t *testing.T) {
	os.Setenv("BUCKET_BASE_URL", "http://test.com/")
	defer os.Unsetenv("BUCKET_BASE_URL")
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/gens?per_page=1&gallery_status=not_submitted", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_role", userrole.RoleNameGALLERY_ADMIN.String())

	MockController.HandleQueryGenerationsForAdmin(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
}

func TestHandleQueryUsersDefaultParams(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryUsers(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var usersResponse repository.UserQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &usersResponse)

	assert.Equal(t, 4, *usersResponse.Total)
	assert.Len(t, usersResponse.Users, 4)
	assert.Nil(t, usersResponse.Next)

	assert.Len(t, usersResponse.TotalByProductID, 1)
	assert.Equal(t, 3, usersResponse.TotalByProductID["prod_123"])

	assert.Equal(t, uuid.MustParse(repository.MOCK_NO_CREDITS_UUID), usersResponse.Users[0].ID)
	assert.Equal(t, "mocknocredituser@stablecog.com", usersResponse.Users[0].Email)
	assert.Equal(t, "4", usersResponse.Users[0].StripeCustomerID)
	assert.Len(t, usersResponse.Users[0].Credits, 0)

	assert.Equal(t, uuid.MustParse(repository.MOCK_ALT_UUID), usersResponse.Users[1].ID)
	assert.Equal(t, "mockaltuser@stablecog.com", usersResponse.Users[1].Email)
	assert.Equal(t, "3", usersResponse.Users[1].StripeCustomerID)
	assert.Len(t, usersResponse.Users[1].Credits, 2)
	assert.Equal(t, int32(100), usersResponse.Users[1].Credits[0].RemainingAmount)
	assert.Equal(t, "mock", usersResponse.Users[1].Credits[0].CreditType.Name)
	assert.Equal(t, int32(1234), usersResponse.Users[1].Credits[1].RemainingAmount)
	assert.Equal(t, "mock", usersResponse.Users[1].Credits[1].CreditType.Name)

	assert.Equal(t, uuid.MustParse(repository.MOCK_NORMAL_UUID), usersResponse.Users[2].ID)
	assert.Equal(t, "mockuser@stablecog.com", usersResponse.Users[2].Email)
	assert.Equal(t, "2", usersResponse.Users[2].StripeCustomerID)
	assert.Len(t, usersResponse.Users[2].Credits, 1)
	assert.Equal(t, int32(100), usersResponse.Users[2].Credits[0].RemainingAmount)
	assert.Equal(t, "mock", usersResponse.Users[2].Credits[0].CreditType.Name)

	assert.Equal(t, uuid.MustParse(repository.MOCK_ADMIN_UUID), usersResponse.Users[3].ID)
	assert.Equal(t, "mockadmin@stablecog.com", usersResponse.Users[3].Email)
	assert.Equal(t, "1", usersResponse.Users[3].StripeCustomerID)
	assert.Len(t, usersResponse.Users[3].Credits, 1)
	assert.Equal(t, int32(100), usersResponse.Users[3].Credits[0].RemainingAmount)
	assert.Equal(t, "prod_123", *usersResponse.Users[3].StripeProductID)
	assert.Equal(t, "mock", usersResponse.Users[3].Credits[0].CreditType.Name)
}

func TestHandleQueryUsersPerPage(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/users?per_page=1", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryUsers(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var usersResponse repository.UserQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &usersResponse)

	assert.Equal(t, 4, *usersResponse.Total)
	assert.Len(t, usersResponse.Users, 1)
	assert.NotNil(t, usersResponse.Next)

	assert.Equal(t, uuid.MustParse(repository.MOCK_NO_CREDITS_UUID), usersResponse.Users[0].ID)
	assert.Equal(t, "mocknocredituser@stablecog.com", usersResponse.Users[0].Email)
	assert.Equal(t, "4", usersResponse.Users[0].StripeCustomerID)
	assert.Len(t, usersResponse.Users[0].Credits, 0)
}

func TestHandleQueryUsersCursor(t *testing.T) {
	w := httptest.NewRecorder()
	// Build request
	req := httptest.NewRequest("GET", "/users?per_page=1", nil)
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryUsers(w, req.WithContext(ctx))
	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	var usersResponse repository.UserQueryMeta
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &usersResponse)

	assert.Equal(t, 4, *usersResponse.Total)
	assert.Len(t, usersResponse.Users, 1)
	assert.NotNil(t, usersResponse.Next)

	offset := *usersResponse.Next

	w = httptest.NewRecorder()
	// Build request
	req = httptest.NewRequest("GET", fmt.Sprintf("/users?per_page=1&cursor=%s", utils.TimeToIsoString(offset)), nil)
	req.Header.Set("Content-Type", "application/json")

	ctx = context.WithValue(req.Context(), "user_id", repository.MOCK_ADMIN_UUID)
	ctx = context.WithValue(ctx, "user_email", repository.MOCK_ADMIN_UUID)

	MockController.HandleQueryUsers(w, req.WithContext(ctx))
	resp = w.Result()
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	respBody, _ = io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &usersResponse)

	assert.Equal(t, 4, *usersResponse.Total)
	assert.Len(t, usersResponse.Users, 1)
	assert.NotNil(t, usersResponse.Next)

	assert.Equal(t, uuid.MustParse(repository.MOCK_ALT_UUID), usersResponse.Users[0].ID)
	assert.Equal(t, "mockaltuser@stablecog.com", usersResponse.Users[0].Email)
	assert.Equal(t, "3", usersResponse.Users[0].StripeCustomerID)
	assert.Len(t, usersResponse.Users[0].Credits, 2)
	assert.Equal(t, int32(100), usersResponse.Users[0].Credits[0].RemainingAmount)
	assert.Equal(t, "mock", usersResponse.Users[0].Credits[0].CreditType.Name)
	assert.Equal(t, int32(1234), usersResponse.Users[0].Credits[1].RemainingAmount)
	assert.Equal(t, "mock", usersResponse.Users[0].Credits[1].CreditType.Name)
}
