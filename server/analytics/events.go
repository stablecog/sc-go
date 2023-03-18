package analytics

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/requests"
)

// Generation | Started
func (a *AnalyticsService) GenerationStarted(user *ent.User, cogReq requests.BaseCogRequest, ip string) error {
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
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Started",
		Properties: properties,
	})
}

// Generation | Succeeded
func (a *AnalyticsService) GenerationSucceeded(user *ent.User, cogReq requests.BaseCogRequest, duration float64, ip string) error {
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
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Duration":          duration,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Succeeded",
		Properties: properties,
	})
}

// Generation | Failed-NSFW
func (a *AnalyticsService) GenerationFailedNSFW(user *ent.User, cogReq requests.BaseCogRequest, duration float64, ip string) error {
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
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Duration":          duration,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Failed-NSFW",
		Properties: properties,
	})
}

// Generation | Failed
func (a *AnalyticsService) GenerationFailed(user *ent.User, cogReq requests.BaseCogRequest, duration float64, failureReason string, ip string) error {
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
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Duration":          duration,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"SC - Failure Reason":    failureReason,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Failed",
		Properties: properties,
	})
}

// Upscale | Started
func (a *AnalyticsService) UpscaleStarted(user *ent.User, cogReq requests.BaseCogRequest, ip string) error {
	// We need to get guidance scale/height/inference steps/width as numeric values
	height, _ := strconv.Atoi(cogReq.Height)
	width, _ := strconv.Atoi(cogReq.Width)

	properties := map[string]interface{}{
		"SC - Height":   height,
		"SC - Width":    width,
		"SC - Model Id": cogReq.ModelId.String(),
		"SC - Scale":    4, // Always 4 for now
		"SC - Image":    cogReq.Image,
		"SC - Type":     cogReq.Type,
		"$ip":           ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Upscale | Started",
		Properties: properties,
	})
}

// Upscale | Succeeded
func (a *AnalyticsService) UpscaleSucceeded(user *ent.User, cogReq requests.BaseCogRequest, duration float64, ip string) error {
	// We need to get guidance scale/height/inference steps/width as numeric values
	height, _ := strconv.Atoi(cogReq.Height)
	width, _ := strconv.Atoi(cogReq.Width)

	properties := map[string]interface{}{
		"SC - Height":   height,
		"SC - Width":    width,
		"SC - Model Id": cogReq.ModelId.String(),
		"SC - Scale":    4, // Always 4 for now
		"SC - Image":    cogReq.Image,
		"SC - Type":     cogReq.Type,
		"SC - Duration": duration,
		"$ip":           ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Upscale | Succeeded",
		Properties: properties,
	})
}

// Upscale | Failed
func (a *AnalyticsService) UpscaleFailed(user *ent.User, cogReq requests.BaseCogRequest, duration float64, failureReason string, ip string) error {
	// We need to get guidance scale/height/inference steps/width as numeric values
	height, _ := strconv.Atoi(cogReq.Height)
	width, _ := strconv.Atoi(cogReq.Width)
	properties := map[string]interface{}{
		"SC - Height":         height,
		"SC - Width":          width,
		"SC - Model Id":       cogReq.ModelId.String(),
		"SC - Scale":          4, // Always 4 for now
		"SC - Image":          cogReq.Image,
		"SC - Type":           cogReq.Type,
		"SC - Duration":       duration,
		"$ip":                 ip,
		"SC - Failure Reason": failureReason,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Upscale | Failed",
		Properties: properties,
	})
}

// Sign Up
func (a *AnalyticsService) SignUp(userId uuid.UUID, email, ipAddress string) error {
	return a.Dispatch(Event{
		DistinctId: userId.String(),
		EventName:  "Sign Up",
		Properties: map[string]interface{}{
			"email":      email,
			"SC - Email": email,
			"$ip":        ipAddress,
		},
	})
}

// New Subscription
func (a *AnalyticsService) Subscription(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id":  productId,
			"SC - Email":              user.Email,
			"SC - Stripe Customer Id": user.StripeCustomerID,
			"$geoip_disable":          true,
		},
	})
}

// Renewed Subscription
func (a *AnalyticsService) SubscriptionRenewal(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription | Renewed",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id":  productId,
			"SC - Email":              user.Email,
			"SC - Stripe Customer Id": user.StripeCustomerID,
			"$geoip_disable":          true,
		},
	})
}

// Cancelled Subscription
func (a *AnalyticsService) SubscriptionCancelled(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription | Cancelled",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id":  productId,
			"SC - Email":              user.Email,
			"SC - Stripe Customer Id": user.StripeCustomerID,
			"$geoip_disable":          true,
		},
	})
}

// Upgraded subscription
func (a *AnalyticsService) SubscriptionUpgraded(user *ent.User, oldProductId string, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription | Upgraded",
		Properties: map[string]interface{}{
			"SC - Stripe Old Product Id": oldProductId,
			"SC - Stripe Product Id":     productId,
			"SC - Email":                 user.Email,
			"SC - Stripe Customer Id":    user.StripeCustomerID,
			"$geoip_disable":             true,
		},
	})
}

// Credit purchase
func (a *AnalyticsService) CreditPurchase(user *ent.User, productId string, amount int) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Credits | Purchased",
		Properties: map[string]interface{}{
			"SC - Stripe Product Id":  productId,
			"SC - Email":              user.Email,
			"SC - Stripe Customer Id": user.StripeCustomerID,
			"SC - Amount":             amount,
			"$geoip_disable":          true,
		},
	})
}

// Free credits replenished
func (a *AnalyticsService) FreeCreditsReplenished(userId uuid.UUID, email string, amount int) error {
	return a.Dispatch(Event{
		DistinctId: userId.String(),
		EventName:  "Credits | Free Replenished",
		Properties: map[string]interface{}{
			"SC - Email":     email,
			"SC - Amount":    amount,
			"$geoip_disable": true,
		},
	})
}
