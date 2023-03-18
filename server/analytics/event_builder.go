package analytics

import (
	"github.com/dukex/mixpanel"
	"github.com/posthog/posthog-go"
)

type Event struct {
	DistinctId string
	EventName  string
	Properties map[string]interface{}
	Identify   bool
}

func (e *Event) PosthogEvent() (posthog.Capture, *posthog.Identify) {
	// Construct properties
	properties := posthog.NewProperties()
	for k, v := range e.Properties {
		properties.Set(k, v)
	}
	c := posthog.Capture{
		DistinctId: e.DistinctId,
		Event:      e.EventName,
		Properties: properties,
	}
	if e.Identify {
		i := posthog.Identify{
			DistinctId: e.DistinctId,
			Properties: properties,
		}
		return c, &i
	}
	return c, nil
}

func (e *Event) MixpanelEvent() (distinctId, eventName string, event *mixpanel.Event, identify *mixpanel.Update) {
	ip := "0"
	// Prune $ip from map if it exists
	mapCopy := make(map[string]interface{})
	for k, v := range e.Properties {
		if k == "$ip" {
			ip = v.(string)
		} else if k == "email" {
			mapCopy["$email"] = v
		} else {
			mapCopy[k] = v
		}
	}
	mixpanelEvent := &mixpanel.Event{
		IP:         ip,
		Properties: mapCopy,
	}
	if e.Identify {
		identify = &mixpanel.Update{
			IP:         ip,
			Properties: mapCopy,
			Operation:  "$set",
		}
	}
	return e.DistinctId, e.EventName, mixpanelEvent, identify
}
