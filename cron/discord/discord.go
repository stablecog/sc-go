package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stablecog/sc-go/cron/models"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Constants
const unhealthyNotificationInterval = 5 * time.Minute
const healthyNotificationInterval = 1 * time.Hour
const rTTL = 2 * time.Hour

type HEALTH_STATUS int

const (
	HEALTHY HEALTH_STATUS = iota
	UNHEALTHY
	UNKNOWN
)

func (h HEALTH_STATUS) StatusString() string {
	if h == HEALTHY {
		return "🟢👌🟢"
	} else if h == UNHEALTHY {
		return "🔴💀🔴"
	}
	return "🟡🤷🟡"
}

// For mocking
var logInfo = log.Info

type DiscordHealthTracker struct {
	ctx                           context.Context
	webhookUrl                    string
	lastStatus                    HEALTH_STATUS
	lastNotificationTime          time.Time
	lastUnhealthyNotificationTime time.Time
	lastHealthyNotificationTime   time.Time
}

// Create new instance of discord health tracker
func NewDiscordHealthTracker(ctx context.Context) *DiscordHealthTracker {
	return &DiscordHealthTracker{
		ctx:        ctx,
		webhookUrl: utils.GetEnv("DISCORD_WEBHOOK_URL", ""),
		// Init last status as UNKNOWN
		lastStatus: UNKNOWN,
	}
}

// Sends a discord notification on either the healthy/unhealthy interval depending on status
func (d *DiscordHealthTracker) SendDiscordNotificationIfNeeded(
	status HEALTH_STATUS,
	generations []*ent.Generation,
	lastGenerationTime time.Time,
) error {
	sinceHealthyNotification := time.Since(d.lastHealthyNotificationTime)
	sinceUnhealthyNotification := time.Since(d.lastUnhealthyNotificationTime)

	// The first time we run (UNKNOWN) we skip notification
	// Otherwise, we sent it if unhealthy and it's been more than unhealthyNotificationInterval
	// Or if healthy and it's been more than healthyNotificationInterval
	if d.lastStatus == UNKNOWN || (status == d.lastStatus &&
		((status == UNHEALTHY && sinceUnhealthyNotification < unhealthyNotificationInterval) ||
			(status == HEALTHY && sinceHealthyNotification < healthyNotificationInterval))) {
		logInfo("Skipping Discord notification, not needed")
		// Set last status
		d.lastStatus = status
		return nil
	}

	start := time.Now().UnixMilli()
	log.Info("Sending Discord notification...")

	// Build webhook body
	webhookBody := getDiscordWebhookBody(status, generations, lastGenerationTime)
	reqBody, err := json.Marshal(webhookBody)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return err
	}
	res, postErr := http.Post(d.webhookUrl, "application/json", bytes.NewBuffer(reqBody))
	if postErr != nil {
		log.Error("Error sending webhook", "err", postErr)
		return postErr
	}
	defer res.Body.Close()

	// Update last notification times
	d.lastNotificationTime = time.Now()
	if status == HEALTHY {
		d.lastHealthyNotificationTime = d.lastNotificationTime
	} else {
		d.lastUnhealthyNotificationTime = d.lastNotificationTime
	}
	end := time.Now().UnixMilli()
	log.Infof("Sent Discord notification in %dms", end-start)

	return nil
}

func getDiscordWebhookBody(
	status HEALTH_STATUS,
	generations []*ent.Generation,
	lastGenerationTime time.Time,
) models.DiscordWebhookBody {
	generationsStr := ""
	generationsStrArr := []string{}
	for _, g := range generations {
		if g.Status == generation.StatusFailed && g.FailureReason != nil && *g.FailureReason == shared.NSFW_ERROR {
			generationsStrArr = append(generationsStrArr, "🌶️")
		} else if g.Status == generation.StatusFailed {
			generationsStrArr = append(generationsStrArr, "🔴")
		} else if g.Status == generation.StatusQueued {
			generationsStrArr = append(generationsStrArr, "⏲️")
		} else if g.Status == generation.StatusStarted {
			generationsStrArr = append(generationsStrArr, "🟡")
		} else {
			generationsStrArr = append(generationsStrArr, "🟢")
		}
	}
	generationsStr = strings.Join(generationsStrArr, "")
	body := models.DiscordWebhookBody{
		Embeds: []models.DiscordWebhookEmbed{
			{
				Color: 11437547,
				Fields: []models.DiscordWebhookField{
					{
						Name:  "Status",
						Value: fmt.Sprintf("```%s```", status.StatusString()),
					},
					{
						Name:  "Generations",
						Value: fmt.Sprintf("```%s```", generationsStr),
					},
					{
						Name:  "Last Generation",
						Value: fmt.Sprintf("```%s```", utils.RelativeTimeStr(lastGenerationTime)),
					},
				},
				Footer: models.DiscordWebhookEmbedFooter{
					Text: time.Now().Format(time.RFC1123),
				},
			},
		},
		Attachments: []models.DiscordWebhookAttachment{},
	}
	return body
}
