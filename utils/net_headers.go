package utils

import (
	"net/http"

	"github.com/mileusna/useragent"
)

// Retrieves a country code via request header, prefer cloudflare, then vercel
func GetCountryCode(r *http.Request) string {
	countryCode := r.Header.Get("CF-IPCountry")
	if countryCode == "" {
		countryCode = r.Header.Get("X-Vercel-IP-Country")
	}
	return countryCode
}

// Parses user agent to return device type, os, and browser
type ClientDeviceType string

const (
	Desktop ClientDeviceType = "desktop"
	Mobile  ClientDeviceType = "mobile"
	Tablet  ClientDeviceType = "tablet"
	Bot     ClientDeviceType = "bot"
	Unknown ClientDeviceType = "unknown"
)

type ClientDeviceInfo struct {
	DeviceType    ClientDeviceType
	DeviceOs      string
	DeviceBrowser string
}

func GetClientDeviceInfo(r *http.Request) ClientDeviceInfo {
	userAgent := r.Header.Get("User-Agent")
	client := useragent.Parse(userAgent)
	deviceType := Unknown
	if client.Mobile {
		deviceType = Mobile
	} else if client.Tablet {
		deviceType = Tablet
	} else if client.Bot {
		deviceType = Bot
	} else if client.Desktop {
		deviceType = Desktop
	}

	return ClientDeviceInfo{
		DeviceType:    deviceType,
		DeviceOs:      client.OS,
		DeviceBrowser: client.Name,
	}
}
