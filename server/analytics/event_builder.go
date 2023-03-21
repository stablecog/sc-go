package analytics

import (
	"github.com/dukex/mixpanel"
	"github.com/posthog/posthog-go"
	"github.com/stablecog/sc-go/shared"
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
	properties.Set("SC - App Version", shared.APP_VERSION)
	c := posthog.Capture{
		DistinctId: e.DistinctId,
		Event:      e.EventName,
		Properties: properties,
	}
	if e.Identify {
		mapOnlyEmail := make(map[string]interface{})
		if email, ok := e.Properties["email"]; ok {
			mapOnlyEmail["email"] = email
			if ip, ok := e.Properties["$ip"]; ok {
				mapOnlyEmail["$ip"] = ip
				i := posthog.Identify{
					DistinctId: e.DistinctId,
					Properties: mapOnlyEmail,
				}
				return c, &i
			}
		}
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
	mapCopy["SC - App Version"] = shared.APP_VERSION
	mixpanelEvent := &mixpanel.Event{
		IP:         ip,
		Properties: mapCopy,
	}
	if e.Identify {
		mapOnlyEmail := make(map[string]interface{})
		if email, ok := mapCopy["$email"]; ok {
			mapOnlyEmail["$email"] = email
			mapOnlyEmail["SC - App Version"] = shared.APP_VERSION
			identify = &mixpanel.Update{
				IP:         ip,
				Properties: mapOnlyEmail,
				Operation:  "$set",
			}
		}
	}
	return e.DistinctId, e.EventName, mixpanelEvent, identify
}
