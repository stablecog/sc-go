package analytics

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/dukex/mixpanel"
	"github.com/hashicorp/go-multierror"
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

	// Setup mixpanel
	mixpanelAPIKey := utils.GetEnv("MIXPANEL_API_KEY", "")
	mixpanelEndpoint := utils.GetEnv("MIXPANEL_ENDPOINT", "")
	if mixpanelAPIKey != "" && mixpanelEndpoint != "" {
		mixpanelClient := mixpanel.New(mixpanelAPIKey, mixpanelEndpoint)
		service.Mixpanel = mixpanelClient
	} else {
		log.Warn("Mixpanel not configured")
	}
	return service
}

func (a *AnalyticsService) Close() {
	if a.Posthog != nil {
		a.Posthog.Close()
	}
}

// Dispatch to all available analytics services
func (a *AnalyticsService) Dispatch(e Event) error {
	var mErr *multierror.Error
	if a.Posthog != nil {
		mErr = multierror.Append(mErr, a.Posthog.Enqueue(e.PosthogEvent()))
	}
	if a.Mixpanel != nil {
		mErr = multierror.Append(mErr, a.Mixpanel.Track(e.MixpanelEvent()))
	}
	err := mErr.ErrorOrNil()
	if err != nil {
		log.Error("Error dispatching analytics event", "err", err)
	}
	return err
}
