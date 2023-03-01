package analytics

import "github.com/xtgo/uuid"

// Sign Up
func (a *AnalyticsService) SignUp(userID uuid.UUID) error {
	return a.Dispatch(Event{
		DistinctId: userID.String(),
		EventName:  "Sign Up",
		Properties: map[string]interface{}{
			"SC - User": userID.String(),
		},
	})
}

// Generation | NSFW
// Subscribe
// Cancelled Subscription
// Downgraded Subscription
// Upgraded Subscription
// Free Credits Replenished
