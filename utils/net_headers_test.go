package utils

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCountryCode(t *testing.T) {
	// 2 methods of getting country code, CF header or vercel header

	// Test that CF is preferred if both present
	request, _ := http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("CF-IPCountry", "cloudflare")
	request.Header.Set("X-Vercel-IP-Country", "vercel")
	assert.Equal(t, "cloudflare", GetCountryCode(request))

	// test that vercel is gotten if cloudflare not presnet
	request, _ = http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("X-Vercel-IP-Country", "vercel")
	assert.Equal(t, "vercel", GetCountryCode(request))

	// Test that empty string if neither provided
	request, _ = http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	assert.Empty(t, GetCountryCode(request))
}

func TestGetClientDeviceInfo(t *testing.T) {
	// Various user agent headers for mobile, tablet, desktop, bot, and unknown
	userAgentDesktop := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
	userAgentMobile := "Mozilla/5.0 (Linux; Android 10; SM-G960U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Mobile Safari/537.36"
	userAgentTablet := "Mozilla/5.0(iPad; U; CPU iPhone OS 3_2 like Mac OS X; en-us) AppleWebKit/531.21.10 (KHTML, like Gecko) Version/4.0.4 Mobile/7B314 Safari/531.21.10"
	userAgentBot := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	userAgentUnknown := "rando user agent lol"

	// Test desktop
	request, _ := http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("User-Agent", userAgentDesktop)
	deviceInfo := GetClientDeviceInfo(request)
	assert.Equal(t, Desktop, deviceInfo.DeviceType)
	assert.Equal(t, "macOS", deviceInfo.DeviceOs)
	assert.Equal(t, "Chrome", deviceInfo.DeviceBrowser)

	// Test mobile
	request, _ = http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("User-Agent", userAgentMobile)
	deviceInfo = GetClientDeviceInfo(request)
	assert.Equal(t, Mobile, deviceInfo.DeviceType)
	assert.Equal(t, "Android", deviceInfo.DeviceOs)
	assert.Equal(t, "Chrome", deviceInfo.DeviceBrowser)

	// Test tablet
	request, _ = http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("User-Agent", userAgentTablet)
	deviceInfo = GetClientDeviceInfo(request)
	assert.Equal(t, Tablet, deviceInfo.DeviceType)
	assert.Equal(t, "iOS", deviceInfo.DeviceOs)
	assert.Equal(t, "Safari", deviceInfo.DeviceBrowser)

	// Test bot
	request, _ = http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("User-Agent", userAgentBot)
	deviceInfo = GetClientDeviceInfo(request)
	assert.Equal(t, Bot, deviceInfo.DeviceType)
	assert.Equal(t, "", deviceInfo.DeviceOs)
	assert.Equal(t, "Googlebot", deviceInfo.DeviceBrowser)

	// Test unknown
	request, _ = http.NewRequest(http.MethodPost, "stablecog.com", bytes.NewReader([]byte("")))
	request.Header.Set("User-Agent", userAgentUnknown)
	deviceInfo = GetClientDeviceInfo(request)
	assert.Equal(t, Unknown, deviceInfo.DeviceType)
	assert.Equal(t, "", deviceInfo.DeviceOs)
	assert.Equal(t, "rando user agent lol", deviceInfo.DeviceBrowser)
}

func TestGetIPAddressFromHeader(t *testing.T) {
	ip := "123.45.67.89"

	// 4 methods of getting IP Address, CF-Connecting-IP preferred, X-Real-Ip, then X-Forwarded-For, then RemoteAddr

	request, _ := http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.Header.Set("CF-Connecting-IP", ip)
	request.Header.Set("X-Real-Ip", "not-the-ip")
	request.Header.Set("X-Forwarded-For", "not-the-ip")
	assert.Equal(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.Header.Set("X-Real-Ip", ip)
	request.Header.Set("X-Forwarded-For", "not-the-ip")

	assert.Equal(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.Header.Set("X-Forwarded-For", ip)
	assert.Equal(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "appditto.com", bytes.NewReader([]byte("")))
	request.RemoteAddr = ip
	assert.Equal(t, ip, GetIPAddress(request))
}
