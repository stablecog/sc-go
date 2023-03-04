package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stablecog/sc-go/cron/models"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

// Sends a discord notification on either the healthy/unhealthy interval depending on status
func FireServerReadyWebhook(version string, msg string) error {
	webhookUrl := utils.GetEnv("DISCORD_WEBHOOK_URL_DEPLOY", "")
	if webhookUrl == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL_DEPLOY not set")
	}
	// Build webhook body
	body := models.DiscordWebhookBody{
		Embeds: []models.DiscordWebhookEmbed{
			{
				Title: fmt.Sprintf(`🟦 Commit: "%s"`, msg),
				Color: 3447003,
				Fields: []models.DiscordWebhookField{
					{
						Value: "```ECS Task Started, Accepting Traffic```",
					},
					{
						// TODO - environment/change
						Name:  "🌏 Environment",
						Value: "```QA```",
					},
				},
				Footer: models.DiscordWebhookEmbedFooter{
					Text: fmt.Sprintf("%s • Version: %s", time.Now().Format(time.RFC1123), version),
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

	return nil
}