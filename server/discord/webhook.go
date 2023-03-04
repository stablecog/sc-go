package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/stablecog/sc-go/cron/models"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

// Sends a discord notification on either the healthy/unhealthy interval depending on status
func FireServerReadyWebhook(version string, msg string, buildStart string) error {
	webhookUrl := utils.GetEnv("DISCORD_WEBHOOK_URL_DEPLOY", "")
	if webhookUrl == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL_DEPLOY not set")
	}
	// Parse build start as int
	buildStartInt, err := strconv.Atoi(buildStart)
	buildStartStr := ""
	if err != nil {
		log.Error("Error parsing build start", "err", err)
	} else {
		buildStartStr = fmt.Sprintf(" in %ds", int(time.Now().Sub(utils.SecondsSinceEpochToTime(int64(buildStartInt))).Seconds()))
	}
	// Build webhook body
	body := models.DiscordWebhookBody{
		Embeds: []models.DiscordWebhookEmbed{
			{
				Title: fmt.Sprintf(`%s  •  %s`, msg, version),
				Color: 5763719,
				Fields: []models.DiscordWebhookField{
					{
						Value: fmt.Sprintf("```Deployed%s```", buildStartStr),
					},
				},
				Footer: models.DiscordWebhookEmbedFooter{
					Text: fmt.Sprintf("%s", time.Now().Format(time.RFC1123)),
				},
			},
		},
		Attachments: []models.DiscordWebhookAttachment{},
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return err
	}
	res, postErr := http.Post(utils.GetEnv("DISCORD_WEBHOOK_URL_DEPLOY", ""), "application/json", bytes.NewBuffer(reqBody))
	if postErr != nil {
		log.Error("Error sending webhook", "err", postErr)
		return postErr
	}
	defer res.Body.Close()

	panic("intentional fail")

	return nil
}
