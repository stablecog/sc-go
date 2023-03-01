package analytics

import (
	"github.com/dukex/mixpanel"
	"github.com/posthog/posthog-go"
)

type Event struct {
	DistinctId string
	EventName  string
	Properties map[string]interface{}
}

func (e *Event) PosthogEvent() posthog.Capture {
	// Construct properties
	properties := posthog.NewProperties()
	for k, v := range e.Properties {
		properties.Set(k, v)
	}
	return posthog.Capture{
		DistinctId: e.DistinctId,
		Event:      e.EventName,
		Properties: properties,
	}
}

func (e *Event) MixpanelEvent() (distinctId, eventName string, event *mixpanel.Event) {
	mixpanelEvent := &mixpanel.Event{
		Properties: e.Properties,
	}
	return e.DistinctId, e.EventName, mixpanelEvent
}
