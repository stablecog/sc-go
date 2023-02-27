package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stablecog/sc-go/cron/models"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

// General redis key prefix
const redisDiscordKeyPrefix = "discord_notification"

// Keep state of health with these keys
var lastHealthyKey = fmt.Sprintf("%s:last_healthy", redisDiscordKeyPrefix)
var lastUnhealthyKey = fmt.Sprintf("%s:last_unhealthy", redisDiscordKeyPrefix)

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
var klogInfof = klog.Infof

type DiscordHealthTracker struct {
	ctx                           context.Context
	webhookUrl                    string
	lastNotificationTime          time.Time
	lastUnhealthyNotificationTime time.Time
	lastHealthyNotificationTime   time.Time
	redis                         *redis.Client
	lastStatus                    HEALTH_STATUS
}

func NewDiscordHealthTracker(ctx context.Context, redis *redis.Client) *DiscordHealthTracker {
	return &DiscordHealthTracker{
		ctx:        ctx,
		webhookUrl: utils.GetEnv("DISCORD_WEBHOOK_URL", ""),
		redis:      redis,
		lastStatus: UNKNOWN,
	}
}

func (d *DiscordHealthTracker) SendDiscordNotificationIfNeeded(
	status HEALTH_STATUS,
	generations []*ent.Generation,
	lastGenerationTime time.Time,
	lastCheckTime time.Time,
) error {
	lastHealthyStr := d.redis.Get(d.ctx, lastHealthyKey).Val()
	lastUnhealthyStr := d.redis.Get(d.ctx, lastUnhealthyKey).Val()
	d.lastHealthyNotificationTime, _ = time.Parse(time.RFC3339, lastHealthyStr)
	d.lastUnhealthyNotificationTime, _ = time.Parse(time.RFC3339, lastUnhealthyStr)

	sinceHealthyNotification := time.Since(d.lastHealthyNotificationTime)
	sinceUnhealthyNotification := time.Since(d.lastUnhealthyNotificationTime)

	if d.lastStatus == UNKNOWN || (status == d.lastStatus &&
		((status == UNHEALTHY && sinceUnhealthyNotification < unhealthyNotificationInterval) ||
			(status == HEALTHY && sinceHealthyNotification < healthyNotificationInterval))) {
		klogInfof("Skipping Discord notification, not needed")
		d.lastStatus = status
		return nil
	}

	start := time.Now().UnixMilli()
	klog.Infoln("Sending Discord notification...")
	webhookBody := getDiscordWebhookBody(status, generations, lastGenerationTime, lastCheckTime)
	reqBody, err := json.Marshal(webhookBody)
	if err != nil {
		klog.Errorf("Error marshalling webhook body: %s", err)
		return err
	}
	res, postErr := http.Post(d.webhookUrl, "application/json", bytes.NewBuffer(reqBody))
	if postErr != nil {
		klog.Errorf("Error sending webhook: %s", postErr)
		return postErr
	}
	defer res.Body.Close()
	d.lastNotificationTime = time.Now()
	if status == HEALTHY {
		err := d.redis.Set(d.ctx, lastHealthyKey, d.lastNotificationTime.Format(time.RFC3339), rTTL).Err()
		if err != nil {
			klog.Error("Redis - Error setting last healthy key: %v", err)
			return err
		}
	} else {
		err := d.redis.Set(d.ctx, lastUnhealthyKey, d.lastNotificationTime.Format(time.RFC3339), rTTL).Err()
		if err != nil {
			klog.Errorf("Redis - Error setting last unhealthy key: %v", err)
			return err
		}
	}
	end := time.Now().UnixMilli()
	klog.Infof("Sent Discord notification in: %dms", end-start)
	return nil
}

func getDiscordWebhookBody(
	status HEALTH_STATUS,
	generations []*ent.Generation,
	lastGenerationTime time.Time,
	lastCheckTime time.Time,
) models.DiscordWebhookBody {
	generationsStr := ""
	generationsStrArr := []string{}
	for _, g := range generations {
		if g.Status == generation.StatusFailed {
			if g.FailureReason != nil && *g.FailureReason == "NSFW" {
				generationsStrArr = append(generationsStrArr, "🌶️")
			} else {
				generationsStrArr = append(generationsStrArr, "🔴")
			}
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
					Text: lastCheckTime.Format(time.RFC1123),
				},
			},
		},
		Attachments: []models.DiscordWebhookAttachment{},
	}
	return body
}
