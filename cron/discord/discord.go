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
const healthyNotificationInterval = 30 * time.Minute
const rTTL = 2 * time.Hour

func StatusString(h shared.HEALTH_STATUS) string {
	if h == shared.HEALTHY {
		return "🟢👌🟢"
	} else if h == shared.UNHEALTHY {
		return "🔴💀🔴"
	}
	return "🟡🤷🟡"
}

// For mocking
var logInfo = log.Info

type DiscordHealthTracker struct {
	ctx        context.Context
	webhookUrl string
	// seenFirstStatus is false until the first check after process start has
	// been observed; that first run has no baseline to compare against, so it
	// never notifies. (lastStatus can't double as this flag anymore since
	// UNKNOWN is now a real, reportable status.)
	seenFirstStatus               bool
	lastStatus                    shared.HEALTH_STATUS
	lastNotificationTime          time.Time
	lastUnhealthyNotificationTime time.Time
	lastHealthyNotificationTime   time.Time
	HTTP                          *http.Client
}

// Create new instance of discord health tracker
func NewDiscordHealthTracker(ctx context.Context) *DiscordHealthTracker {
	return &DiscordHealthTracker{
		ctx:        ctx,
		webhookUrl: utils.GetEnv().DiscordWebhookUrl,
		// Init last status as UNKNOWN
		lastStatus: shared.UNKNOWN,
		HTTP:       &http.Client{},
	}
}

// Sends a discord notification on either the healthy/unhealthy interval depending on status.
// `asOf` is the reference time used to format relative timestamps and the embed footer — it
// should be the moment the input data was sampled, NOT the moment the notification is being
// sent. Otherwise, a slow health check (e.g. one that spent minutes in test-generation
// retries) would render its old snapshot against `time.Now()` and produce timestamps like
// "20m ago" right next to a fresh run reporting "Just now" — looking like an impossible
// flip-flop in the channel.
// `healthErr` is the reason behind a non-HEALTHY status (test generation error,
// traffic failure rate); it is rendered as its own embed field so the alert
// itself says what went wrong.
func (d *DiscordHealthTracker) SendDiscordNotificationIfNeeded(
	status shared.HEALTH_STATUS,
	generations []*ent.Generation,
	lastGenerationTime time.Time,
	lastSuccessfulGenerationTime time.Time,
	isRunpodServerlessActive bool,
	runpodServerlessErr error,
	healthErr error,
	asOf time.Time,
) error {
	sinceHealthyNotification := time.Since(d.lastHealthyNotificationTime)
	sinceUnhealthyNotification := time.Since(d.lastUnhealthyNotificationTime)

	shouldSkip := false
	statusUnchanged := status == d.lastStatus

	// The first time we run we skip notification
	if !d.seenFirstStatus {
		shouldSkip = true
		d.seenFirstStatus = true
	}

	// If status didn't change and healthy notification interval hasn't passed, skip
	if statusUnchanged && status == shared.HEALTHY && sinceHealthyNotification < healthyNotificationInterval {
		shouldSkip = true
	}

	// If status didn't change and the unhealthy notification interval hasn't
	// passed, skip. UNKNOWN (broken health check: bad key, no credits, ...)
	// uses the same cadence — it needs human attention just like UNHEALTHY.
	if statusUnchanged && status != shared.HEALTHY && sinceUnhealthyNotification < unhealthyNotificationInterval {
		shouldSkip = true
	}

	d.lastStatus = status

	if shouldSkip {
		logInfo("Skipping Discord notification, not needed")
		return nil
	}

	start := time.Now().UnixMilli()
	log.Info("Sending Discord notification...")

	// Build webhook body
	webhookBody := getDiscordWebhookBody(
		status,
		generations,
		lastGenerationTime,
		lastSuccessfulGenerationTime,
		isRunpodServerlessActive,
		runpodServerlessErr,
		healthErr,
		asOf,
	)
	reqBody, err := json.Marshal(webhookBody)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return err
	}

	req, err := http.NewRequest("POST", d.webhookUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Error("Error creating request", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := d.HTTP.Do(req)
	if err != nil {
		log.Error("Error sending webhook", "err", err)
		return err
	}
	defer res.Body.Close()

	// Update last notification times
	d.lastNotificationTime = time.Now()
	if status == shared.HEALTHY {
		d.lastHealthyNotificationTime = d.lastNotificationTime
	} else {
		d.lastUnhealthyNotificationTime = d.lastNotificationTime
	}
	end := time.Now().UnixMilli()
	log.Infof("Sent Discord notification in %dms", end-start)

	return nil
}

func getDiscordWebhookBody(
	status shared.HEALTH_STATUS,
	generations []*ent.Generation,
	lastGenerationTime time.Time,
	lastSuccessfulGenerationTime time.Time,
	isRunpodServerlessActive bool,
	runpodServerlessErr error,
	healthErr error,
	asOf time.Time,
) models.DiscordWebhookBody {
	generationsStr := ""
	generationsStrArr := []string{}

	discordUserIds := utils.GetEnv().GetDiscordUserIdsToNotify()

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

	var content *string
	if status != shared.HEALTHY && len(discordUserIds) > 0 {
		mentionStr := ""
		for _, userId := range discordUserIds {
			mentionStr += fmt.Sprintf("<@%s> ", userId)
		}
		content = &mentionStr
	}

	isRunpodServerlessActiveStr := "⚪️ Inactive"

	if isRunpodServerlessActive {
		isRunpodServerlessActiveStr = "🟢 Active"
	}

	body := models.DiscordWebhookBody{
		Content: content,
		Embeds: []models.DiscordWebhookEmbed{
			{
				Color: 11437547,
				Fields: []models.DiscordWebhookField{
					{
						Name:  "Status",
						Value: fmt.Sprintf("```%s```", StatusString(status)),
					},
					{
						Name:  "Generations",
						Value: fmt.Sprintf("```%s```", generationsStr),
					},
					{
						Name:  "Last Generation",
						Value: fmt.Sprintf("```%s```", utils.RelativeTimeStrFrom(lastGenerationTime, asOf)),
					},
					{
						Name:  "Last Successful Generation",
						Value: fmt.Sprintf("```%s```", utils.RelativeTimeStrFrom(lastSuccessfulGenerationTime, asOf)),
					},
					{
						Name:  "Runpod Serverless Status",
						Value: fmt.Sprintf("```%s```", isRunpodServerlessActiveStr),
					},
				},
				Footer: models.DiscordWebhookEmbedFooter{
					Text: asOf.Format(time.RFC1123),
				},
			},
		},
		Attachments: []models.DiscordWebhookAttachment{},
	}

	if healthErr != nil {
		body.Embeds[0].Fields = append(body.Embeds[0].Fields, models.DiscordWebhookField{
			Name:  "Health Check Error",
			Value: fmt.Sprintf("```%s```", healthErr.Error()),
		})
	}

	if runpodServerlessErr != nil {
		body.Embeds[0].Fields = append(body.Embeds[0].Fields, models.DiscordWebhookField{
			Name:  "Runpod Serverless Error",
			Value: fmt.Sprintf("```%s```", runpodServerlessErr.Error()),
		})
	}

	return body
}
