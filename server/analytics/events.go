package analytics

import (
	"strconv"

	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/requests"
)

// Sign Up
func (a *AnalyticsService) GenerationStarted(user *ent.User, cogReq requests.BaseCogRequest) error {
	// We need to get guidance scale/height/inference steps/width as numeric values
	height, _ := strconv.Atoi(cogReq.Height)
	width, _ := strconv.Atoi(cogReq.Width)
	inferenceSteps, _ := strconv.Atoi(cogReq.NumInferenceSteps)
	// Guidance scale is a float
	guidanceScale, _ := strconv.ParseFloat(cogReq.GuidanceScale, 32)

	properties := map[string]interface{}{
		"SC - Guidance Scale":    guidanceScale,
		"SC - Height":            height,
		"SC - Width":             width,
		"SC - Inference Steps":   inferenceSteps,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Scheduler Id":      cogReq.SchedulerId.String(),
		"SC - Product Id":        user.ActiveProductID,
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
	}
	if user.ActiveProductID != nil {
		properties["SC - Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Started",
		Properties: properties,
	})
}

// Generation | NSFW
// Subscribe
// Cancelled Subscription
// Downgraded Subscription
// Upgraded Subscription
// Free Credits Replenished
