package requests

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Parse query filters from url params
func TestParseQueryGenerationFilters(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&upscaled=only&start_dt=2021-01-01T00:00:00Z&gallery_status=approved,not_submitted&order_by=updated_at&is_favorited=true&was_auto_submitted=true"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	filters := &QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(values)
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
	assert.Equal(t, UpscaleStatusOnly, filters.UpscaleStatus)
	assert.NotNil(t, filters.StartDt)
	assert.Equal(t, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), *filters.StartDt)
	assert.Len(t, filters.GalleryStatus, 2)
	assert.Equal(t, OrderByUpdatedAt, filters.OrderBy)
	assert.True(t, *filters.IsFavorited)
	assert.True(t, *filters.WasAutoSubmitted)
	// Default descending
	assert.Equal(t, SortOrderDescending, filters.Order)

	// Check is_favorited since it can be nil
	urlStr = "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&upscaled=only&start_dt=2021-01-01T00:00:00Z&gallery_status=approved,not_submitted&order_by=updated_at"
	// Get url.Values from string
	values, err = url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	filters = &QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(values)
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
	assert.Equal(t, UpscaleStatusOnly, filters.UpscaleStatus)
	assert.NotNil(t, filters.StartDt)
	assert.Equal(t, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), *filters.StartDt)
	assert.Len(t, filters.GalleryStatus, 2)
	assert.Equal(t, OrderByUpdatedAt, filters.OrderBy)
	assert.Nil(t, filters.IsFavorited)
	// Default descending
	assert.Equal(t, SortOrderDescending, filters.Order)
}

func TestParseQueryGenerationFilterError(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&order=invalid"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	filters := &QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(values)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid order: 'invalid' expected 'asc' or 'desc'", err.Error())
}

func TestToQdrantFilters(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&upscaled=only&start_dt=2021-01-01T00:00:00Z&gallery_status=approved,not_submitted&order_by=updated_at&is_favorited=true&was_auto_submitted=true"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	filters := &QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(values)
	assert.Nil(t, err)

	// Convert to Qdrant filters
	qdrantFilters, _ := filters.ToQdrantFilters(false)
	// marshal
	b, err := json.Marshal(qdrantFilters)
	assert.Nil(t, err)

	// assert equal to prefined string
	assert.Equal(t, "{\"must\":[{\"key\":\"model\",\"match\":{\"value\":\"49d75ae2-5407-40d9-8c02-0c44ba08f358\"}},{\"key\":\"scheduler\",\"match\":{\"value\":\"e07ad712-41ad-4ff7-8727-faf0d91e4c4e\"}},{\"key\":\"scheduler\",\"match\":{\"value\":\"c09aaf4d-2d78-4281-89aa-88d5d0a5d70b\"}},{\"key\":\"height\",\"range\":{\"gte\":6}},{\"key\":\"height\",\"range\":{\"lte\":7}},{\"key\":\"width\",\"range\":{\"gte\":1}},{\"key\":\"width\",\"range\":{\"lte\":5}},{\"key\":\"inference_steps\",\"range\":{\"gte\":2}},{\"key\":\"inference_steps\",\"range\":{\"lte\":3}},{\"key\":\"guidance_scale\",\"range\":{\"gte\":2}},{\"key\":\"width\",\"range\":{\"lte\":4}},{\"key\":\"gallery_status\",\"match\":{\"value\":\"approved\"}},{\"key\":\"gallery_status\",\"match\":{\"value\":\"not_submitted\"}},{\"key\":\"created_at\",\"range\":{\"gte\":1609459200}},{\"key\":\"is_favorited\",\"match\":{\"value\":true}},{\"key\":\"was_auto_submitted\",\"match\":{\"value\":true}}],\"must_not\":[{\"is_empty\":{\"key\":\"upscaled_image_path\"}}],\"should\":[{\"key\":\"height\",\"match\":{\"value\":512}},{\"key\":\"width\",\"match\":{\"value\":512}},{\"key\":\"width\",\"match\":{\"value\":768}},{\"key\":\"inference_steps\",\"match\":{\"value\":30}},{\"key\":\"guidance_scale\",\"match\":{\"value\":5}}]}", string(b))
}

// Ignore gallery status filters
func TestToQdrantFiltersIgnoreGalleryStatus(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&upscaled=only&start_dt=2021-01-01T00:00:00Z&gallery_status=approved,not_submitted&order_by=updated_at&is_favorited=true&was_auto_submitted=true"
	// Get url.Values from string
	values, err := url.ParseQuery(urlStr)
	assert.Nil(t, err)
	// Parse filters
	filters := &QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(values)
	assert.Nil(t, err)

	// Convert to Qdrant filters
	qdrantFilters, _ := filters.ToQdrantFilters(true)
	// marshal
	b, err := json.Marshal(qdrantFilters)
	assert.Nil(t, err)

	// assert equal to prefined string
	assert.Equal(t, "{\"must\":[{\"key\":\"model\",\"match\":{\"value\":\"49d75ae2-5407-40d9-8c02-0c44ba08f358\"}},{\"key\":\"scheduler\",\"match\":{\"value\":\"e07ad712-41ad-4ff7-8727-faf0d91e4c4e\"}},{\"key\":\"scheduler\",\"match\":{\"value\":\"c09aaf4d-2d78-4281-89aa-88d5d0a5d70b\"}},{\"key\":\"height\",\"range\":{\"gte\":6}},{\"key\":\"height\",\"range\":{\"lte\":7}},{\"key\":\"width\",\"range\":{\"gte\":1}},{\"key\":\"width\",\"range\":{\"lte\":5}},{\"key\":\"inference_steps\",\"range\":{\"gte\":2}},{\"key\":\"inference_steps\",\"range\":{\"lte\":3}},{\"key\":\"guidance_scale\",\"range\":{\"gte\":2}},{\"key\":\"width\",\"range\":{\"lte\":4}},{\"key\":\"created_at\",\"range\":{\"gte\":1609459200}},{\"key\":\"is_favorited\",\"match\":{\"value\":true}},{\"key\":\"was_auto_submitted\",\"match\":{\"value\":true}}],\"must_not\":[{\"is_empty\":{\"key\":\"upscaled_image_path\"}}],\"should\":[{\"key\":\"height\",\"match\":{\"value\":512}},{\"key\":\"width\",\"match\":{\"value\":512}},{\"key\":\"width\",\"match\":{\"value\":768}},{\"key\":\"inference_steps\",\"match\":{\"value\":30}},{\"key\":\"guidance_scale\",\"match\":{\"value\":5}}]}", string(b))
}
