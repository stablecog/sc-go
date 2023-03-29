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
	skipIdentify := false
	// Construct properties
	properties := posthog.NewProperties()
	for k, v := range e.Properties {
		if k == "$ip" && v == "system" {
			skipIdentify = true
			continue
		}
		properties.Set(k, v)
	}
	properties.Set("SC - App Version", shared.APP_VERSION)
	c := posthog.Capture{
		DistinctId: e.DistinctId,
		Event:      e.EventName,
		Properties: properties,
	}
	if e.Identify && !skipIdentify {
		propertiesIdentify := posthog.NewProperties()
		// Remove all properites except email, app version, device_type/browser/os/version
		for k := range e.Properties {
			if k != "email" && k != "SC - App Version" && k != "$device_type" && k != "$browser" && k != "$os" && k != "$browser_version" && k != "$ip" {
				continue
			}
			propertiesIdentify[k] = e.Properties[k]
		}
		i := posthog.Identify{
			DistinctId: e.DistinctId,
			Properties: propertiesIdentify,
		}
		return c, &i
	}
	return c, nil
}

func (e *Event) MixpanelEvent() (distinctId, eventName string, event *mixpanel.Event, identify *mixpanel.Update) {
	ip := "0"
	skipIdentify := false
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
	if ip == "system" {
		skipIdentify = true
		ip = ""
	}
	mapCopy["SC - App Version"] = shared.APP_VERSION
	mixpanelEvent := &mixpanel.Event{
		IP:         ip,
		Properties: mapCopy,
	}
	if e.Identify && !skipIdentify {
		mapOnlyEmail := make(map[string]interface{})
		// Remove all properites except email, app version, device_type/browser/os/version
		for k := range mapCopy {
			if k != "$email" && k != "SC - App Version" && k != "$device_type" && k != "$browser" && k != "$os" && k != "$browser_version" {
				continue
			}
			mapOnlyEmail[k] = mapCopy[k]
		}
		identify = &mixpanel.Update{
			IP:         ip,
			Properties: mapOnlyEmail,
			Operation:  "$set",
		}
	}
	return e.DistinctId, e.EventName, mixpanelEvent, identify
}
