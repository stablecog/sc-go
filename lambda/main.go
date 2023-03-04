package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stablecog/sc-go/cron/models"
	"github.com/stablecog/sc-go/log"
)

var WebhookUrl string

// Sends a discord notification on either the healthy/unhealthy interval depending on status
func FireWebhook(data events.SNSEvent) error {
	// Encode to json string
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return err
	}
	// Build webhook body
	body := models.DiscordWebhookBody{
		Embeds: []models.DiscordWebhookEmbed{
			{
				Title: "ECS Error Event",
				Color: 15548997,
				Fields: []models.DiscordWebhookField{
					{
						Value: fmt.Sprintf(`%s`, string(jsonData)),
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
	res, postErr := http.Post(WebhookUrl, "application/json", bytes.NewBuffer(reqBody))
	if postErr != nil {
		log.Error("Error sending webhook", "err", postErr)
		return postErr
	}
	defer res.Body.Close()

	return nil
}

func HandleRequest(ctx context.Context, event events.SNSEvent) (string, error) {
	err := FireWebhook(event)
	return "", err
}

func main() {
	lambda.Start(HandleRequest)
}
