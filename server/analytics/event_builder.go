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
	ip := "0"
	// Prune $ip from map if it exists
	mapCopy := make(map[string]interface{})
	for k, v := range e.Properties {
		if k == "$ip" {
			ip = v.(string)
		} else {
			mapCopy[k] = v
		}
	}
	mixpanelEvent := &mixpanel.Event{
		IP:         ip,
		Properties: mapCopy,
	}
	return e.DistinctId, e.EventName, mixpanelEvent
}
