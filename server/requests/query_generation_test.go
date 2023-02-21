package requests

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Parse query filters from url params
func TestParseQueryGenerationFilters(t *testing.T) {
	urlStr := "/gens?per_page=1&cursor=2021-01-01T00:00:00Z&min_width=1&max_width=5&min_height=6&max_height=7&max_inference_steps=3&min_inference_steps=2&max_guidance_scale=4&min_guidance_scale=2&widths=512,768&heights=512&inference_steps=30&guidance_scales=5&scheduler_ids=e07ad712-41ad-4ff7-8727-faf0d91e4c4e,c09aaf4d-2d78-4281-89aa-88d5d0a5d70b&model_ids=49d75ae2-5407-40d9-8c02-0c44ba08f358&upscaled=only&start_dt=2021-01-01T00:00:00Z&gallery_status=accepted,not_submitted"
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
