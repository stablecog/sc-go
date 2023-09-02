package analytics

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

func setDeviceInfo(dInfo utils.ClientDeviceInfo, properties map[string]interface{}) {
	if dInfo.DeviceBrowser != "" {
		properties["$browser"] = dInfo.DeviceBrowser
	}
	if dInfo.DeviceOs != "" {
		properties["$os"] = dInfo.DeviceOs
	}
	if dInfo.DeviceBrowserVersion != "" {
		properties["$browser_version"] = dInfo.DeviceBrowserVersion
	}
	if dInfo.DeviceType != "" {
		properties["$device_type"] = dInfo.DeviceType
	}
}

// Generation | Started
func (a *AnalyticsService) GenerationStarted(user *ent.User, cogReq requests.BaseCogRequest, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Guidance Scale":    *cogReq.GuidanceScale,
		"SC - Height":            *cogReq.Height,
		"SC - Width":             *cogReq.Width,
		"SC - Inference Steps":   *cogReq.NumInferenceSteps,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Scheduler Id":      cogReq.SchedulerId.String(),
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"SC - Source":            source,
		"$ip":                    ip,
		"email":                  user.Email,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	if cogReq.InitImageUrlS3 != "" {
		properties["SC - Init Image Url"] = cogReq.InitImageUrlS3
	}
	if cogReq.PromptStrength != nil {
		properties["SC - Prompt Strength"] = *cogReq.PromptStrength
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		Identify:   true,
		DistinctId: user.ID.String(),
		EventName:  "Generation | Started",
		Properties: properties,
	})
}

// Generation | Succeeded
func (a *AnalyticsService) GenerationSucceeded(user *ent.User, cogReq requests.BaseCogRequest, duration float64, qDuration float64, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Guidance Scale":    *cogReq.GuidanceScale,
		"SC - Height":            *cogReq.Height,
		"SC - Width":             *cogReq.Width,
		"SC - Inference Steps":   *cogReq.NumInferenceSteps,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Scheduler Id":      cogReq.SchedulerId.String(),
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Duration":          duration,
		"SC - Duration in Queue": qDuration,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"SC - Source":            source,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	if cogReq.InitImageUrlS3 != "" {
		properties["SC - Init Image Url"] = cogReq.InitImageUrlS3
	}
	if cogReq.PromptStrength != nil {
		properties["SC - Prompt Strength"] = *cogReq.PromptStrength
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Succeeded",
		Properties: properties,
	})
}

// Generation | Failed-NSFW
func (a *AnalyticsService) GenerationFailedNSFW(user *ent.User, cogReq requests.BaseCogRequest, duration float64, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Guidance Scale":    *cogReq.GuidanceScale,
		"SC - Height":            *cogReq.Height,
		"SC - Width":             *cogReq.Width,
		"SC - Inference Steps":   *cogReq.NumInferenceSteps,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Scheduler Id":      cogReq.SchedulerId.String(),
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Duration":          duration,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"SC - Source":            source,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	if cogReq.InitImageUrlS3 != "" {
		properties["SC - Init Image Url"] = cogReq.InitImageUrlS3
	}
	if cogReq.PromptStrength != nil {
		properties["SC - Prompt Strength"] = *cogReq.PromptStrength
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Failed-NSFW",
		Properties: properties,
	})
}

// Generation | Failed
func (a *AnalyticsService) GenerationFailed(user *ent.User, cogReq requests.BaseCogRequest, duration float64, failureReason string, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Guidance Scale":    *cogReq.GuidanceScale,
		"SC - Height":            *cogReq.Height,
		"SC - Width":             *cogReq.Width,
		"SC - Inference Steps":   *cogReq.NumInferenceSteps,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Scheduler Id":      cogReq.SchedulerId.String(),
		"SC - Submit to Gallery": cogReq.SubmitToGallery,
		"SC - Duration":          duration,
		"SC - Num Outputs":       cogReq.NumOutputs,
		"SC - Failure Reason":    failureReason,
		"SC - Source":            source,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	if cogReq.InitImageUrlS3 != "" {
		properties["SC - Init Image Url"] = cogReq.InitImageUrlS3
	}
	if cogReq.PromptStrength != nil {
		properties["SC - Prompt Strength"] = *cogReq.PromptStrength
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Failed",
		Properties: properties,
	})
}

// Upscale | Started
func (a *AnalyticsService) UpscaleStarted(user *ent.User, cogReq requests.BaseCogRequest, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":  user.ID,
		"SC - Height":   *cogReq.Height,
		"SC - Width":    *cogReq.Width,
		"SC - Model Id": cogReq.ModelId.String(),
		"SC - Scale":    4, // Always 4 for now
		"SC - Image":    cogReq.Image,
		"SC - Type":     cogReq.Type,
		"SC - Source":   source,
		"$ip":           ip,
		"email":         user.Email,
	}
	if ip == "system" {
		properties["SC - System Generated"] = true
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		Identify:   true,
		DistinctId: user.ID.String(),
		EventName:  "Upscale | Started",
		Properties: properties,
	})
}

// Upscale | Succeeded
func (a *AnalyticsService) UpscaleSucceeded(user *ent.User, cogReq requests.BaseCogRequest, duration float64, qDuration float64, source enttypes.SourceType, ip string) error {

	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Height":            *cogReq.Height,
		"SC - Width":             *cogReq.Width,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Scale":             4, // Always 4 for now
		"SC - Image":             cogReq.Image,
		"SC - Type":              cogReq.Type,
		"SC - Duration":          duration,
		"SC - Duration in Queue": qDuration,
		"SC - Source":            source,
		"$ip":                    ip,
	}
	if ip == "system" {
		properties["SC - System Generated"] = true
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Upscale | Succeeded",
		Properties: properties,
	})
}

// Upscale | Failed
func (a *AnalyticsService) UpscaleFailed(user *ent.User, cogReq requests.BaseCogRequest, duration float64, failureReason string, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":        user.ID,
		"SC - Height":         *cogReq.Height,
		"SC - Width":          *cogReq.Width,
		"SC - Model Id":       cogReq.ModelId.String(),
		"SC - Scale":          4, // Always 4 for now
		"SC - Image":          cogReq.Image,
		"SC - Type":           cogReq.Type,
		"SC - Duration":       duration,
		"SC - Source":         source,
		"$ip":                 ip,
		"SC - Failure Reason": failureReason,
	}
	if ip == "system" {
		properties["SC - System Generated"] = true
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Upscale | Failed",
		Properties: properties,
	})
}

// Voiceover | Started
func (a *AnalyticsService) VoiceoverStarted(user *ent.User, cogReq requests.BaseCogRequest, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":        user.ID,
		"SC - Model Id":       cogReq.ModelId.String(),
		"SC - Speaker Id":     cogReq.SpeakerId.String(),
		"SC - Temperature":    *cogReq.Temp,
		"SC - Denoise Audio":  *cogReq.DenoiseAudio,
		"SC - Remove Silence": *cogReq.RemoveSilence,
		"SC - Source":         source,
		"$ip":                 ip,
		"email":               user.Email,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		Identify:   true,
		DistinctId: user.ID.String(),
		EventName:  "Voiceover | Started",
		Properties: properties,
	})
}

// Voiceover | Succeeded
func (a *AnalyticsService) VoiceoverSucceeded(user *ent.User, cogReq requests.BaseCogRequest, duration float64, qDuration float64, source enttypes.SourceType, ip string) error {

	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Model Id":          cogReq.ModelId.String(),
		"SC - Speaker Id":        cogReq.SpeakerId.String(),
		"SC - Temperature":       *cogReq.Temp,
		"SC - Denoise Audio":     *cogReq.DenoiseAudio,
		"SC - Remove Silence":    *cogReq.RemoveSilence,
		"SC - Duration":          duration,
		"SC - Duration in Queue": qDuration,
		"SC - Source":            source,
		"$ip":                    ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Voiceover | Succeeded",
		Properties: properties,
	})
}

// Voiceover | Failed
func (a *AnalyticsService) VoiceoverFailed(user *ent.User, cogReq requests.BaseCogRequest, duration float64, failureReason string, source enttypes.SourceType, ip string) error {
	properties := map[string]interface{}{
		"SC - User Id":        user.ID,
		"SC - Model Id":       cogReq.ModelId.String(),
		"SC - Speaker Id":     cogReq.SpeakerId.String(),
		"SC - Temperature":    *cogReq.Temp,
		"SC - Denoise Audio":  *cogReq.DenoiseAudio,
		"SC - Remove Silence": *cogReq.RemoveSilence,
		"SC - Duration":       duration,
		"$ip":                 ip,
		"SC - Failure Reason": failureReason,
		"SC - Source":         source,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Voiceover | Failed",
		Properties: properties,
	})
}

// Sign Up
func (a *AnalyticsService) SignUp(userId uuid.UUID, email, ipAddress string, deviceInfo utils.ClientDeviceInfo) error {
	properties := map[string]interface{}{
		"email":        email,
		"SC - Email":   email,
		"SC - User Id": userId,
		"$ip":          ipAddress,
	}
	setDeviceInfo(deviceInfo, properties)
	return a.Dispatch(Event{
		Identify:   true,
		DistinctId: userId.String(),
		EventName:  "Sign Up",
		Properties: properties,
	})
}

// New Subscription
func (a *AnalyticsService) Subscription(user *ent.User, productId string) error {
	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Subscription",
		Properties: map[string]interface{}{
			"SC - User Id":            user.ID,
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
			"SC - User Id":            user.ID,
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
			"SC - User Id":            user.ID,
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
			"SC - User Id":               user.ID,
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
			"SC - User Id":            user.ID,
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
			"SC - User Id":   userId,
			"SC - Email":     email,
			"SC - Amount":    amount,
			"$geoip_disable": true,
		},
	})
}

func (a *AnalyticsService) GenerationFailedNSFWPrompt(
	user *ent.User,
	cogReq requests.BaseCogRequest,
	failureSource string,
	source enttypes.SourceType,
	translatedPrompt string,
	similarToBannedPromptId string,
	similarityScore float64,
	moderationAPIReason string,
	moderationAPIScore float32,
	ip string,
) error {
	properties := map[string]interface{}{
		"SC - User Id":           user.ID,
		"SC - Failure Source":    failureSource,
		"SC - Source":            source,
		"SC - Original Prompt":   cogReq.Prompt,
		"SC - Translated Prompt": translatedPrompt,
		"$ip":                    ip,
	}
	if similarToBannedPromptId != "" {
		properties["SC - Similar to Banned Prompt Id"] = similarToBannedPromptId
	}
	if similarityScore > 0 {
		properties["SC - Similarity Score"] = similarityScore
	}
	if moderationAPIReason != "" {
		properties["SC - Moderation API Reason"] = moderationAPIReason
	}
	if moderationAPIScore > 0 {
		properties["SC - Moderation API Score"] = moderationAPIScore
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Generation | Failed (NSFW Prompt)",
		Properties: properties,
	})
}

func (a *AnalyticsService) AutoBannedForBannedPromptEmbeddingViolation(
	user *ent.User,
	cogReq requests.BaseCogRequest,
	source enttypes.SourceType,
	translatedPrompt string,
	similarToBannedPromptId string,
	similarityScore float64,
	violationCount int,
	ip string,
) error {
	properties := map[string]interface{}{
		"SC - User Id":                     user.ID,
		"SC - Source":                      source,
		"SC - Original Prompt":             cogReq.Prompt,
		"SC - Translated Prompt":           translatedPrompt,
		"SC - Violation Count":             violationCount,
		"SC - Similar to Banned Prompt Id": similarToBannedPromptId,
		"SC - Similarity Score":            similarityScore,
		"$ip":                              ip,
	}
	if user.ActiveProductID != nil {
		properties["SC - Stripe Product Id"] = user.ActiveProductID
	}
	setDeviceInfo(cogReq.DeviceInfo, properties)

	return a.Dispatch(Event{
		DistinctId: user.ID.String(),
		EventName:  "Auto Banned | Banned Prompt Embedding Violation",
		Properties: properties,
	})
}
