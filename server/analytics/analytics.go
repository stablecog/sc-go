package analytics

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/dukex/mixpanel"
	"github.com/posthog/posthog-go"
	"github.com/stablecog/sc-go/utils"
)

type AnalyticsService struct {
	Posthog  posthog.Client
	Mixpanel mixpanel.Mixpanel
}

func NewAnalyticsService() *AnalyticsService {
	service := &AnalyticsService{}
	// Setup posthog
	posthogAPIKey := utils.GetEnv("POSTHOG_API_KEY", "")
	posthogEndpoint := utils.GetEnv("POSTHOG_ENDPOINT", "")
	if posthogAPIKey != "" && posthogEndpoint != "" {
		client, err := posthog.NewWithConfig(
			posthogAPIKey,
			posthog.Config{
				Endpoint: posthogEndpoint,
			},
		)
		if err != nil {
			log.Fatal("Error connecting to posthog", "err", err)
			os.Exit(1)
		}
		service.Posthog = client
	}

	return service
}

// Dispatch to all available analytics services
func (a *AnalyticsService) Dispatch(e Event) (err error) {
	if a.Posthog != nil {
		err = a.Posthog.Enqueue(e.PosthogEvent())
	}
	if err != nil {
		log.Error("Error dispatching analytics event", "err", err)
	}
	return err
}
