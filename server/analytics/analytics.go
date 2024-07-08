package analytics

import (
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/posthog/posthog-go"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

type AnalyticsService struct {
	Posthog posthog.Client
}

func NewAnalyticsService() *AnalyticsService {
	service := &AnalyticsService{}
	// Setup posthog
	posthogAPIKey := utils.GetEnv().PosthogApiKey
	posthogEndpoint := utils.GetEnv().PosthogEndpoint
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
	} else {
		log.Warn("Posthog not configured")
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
		capture, identify := e.PosthogEvent()
		if identify != nil {
			mErr = multierror.Append(mErr, a.Posthog.Enqueue(*identify))
		}
		mErr = multierror.Append(mErr, a.Posthog.Enqueue(capture))
	}
	err := mErr.ErrorOrNil()
	if err != nil {
		log.Error("Error dispatching analytics event", "err", err)
	}
	return err
}
