package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stablecog/sc-go/cron/models"
	"github.com/stablecog/sc-go/log"
)

var WebhookUrl string

type SNSMessage struct {
	AlarmName                          string        `json:"AlarmName"`
	AlarmDescription                   interface{}   `json:"AlarmDescription"`
	AWSAccountID                       string        `json:"AWSAccountId"`
	AlarmConfigurationUpdatedTimestamp string        `json:"AlarmConfigurationUpdatedTimestamp"`
	NewStateValue                      string        `json:"NewStateValue"`
	NewStateReason                     string        `json:"NewStateReason"`
	StateChangeTime                    string        `json:"StateChangeTime"`
	Region                             string        `json:"Region"`
	AlarmArn                           string        `json:"AlarmArn"`
	OldStateValue                      string        `json:"OldStateValue"`
	OKActions                          []interface{} `json:"OKActions"`
	AlarmActions                       []string      `json:"AlarmActions"`
	InsufficientDataActions            []interface{} `json:"InsufficientDataActions"`
	Trigger                            struct {
		MetricName                       string        `json:"MetricName"`
		Namespace                        string        `json:"Namespace"`
		StatisticType                    string        `json:"StatisticType"`
		Statistic                        string        `json:"Statistic"`
		Unit                             interface{}   `json:"Unit"`
		Dimensions                       []interface{} `json:"Dimensions"`
		Period                           int           `json:"Period"`
		EvaluationPeriods                int           `json:"EvaluationPeriods"`
		DatapointsToAlarm                int           `json:"DatapointsToAlarm"`
		ComparisonOperator               string        `json:"ComparisonOperator"`
		Threshold                        float64       `json:"Threshold"`
		TreatMissingData                 string        `json:"TreatMissingData"`
		EvaluateLowSampleCountPercentile string        `json:"EvaluateLowSampleCountPercentile"`
	} `json:"Trigger"`
}

// Sends a discord notification on either the healthy/unhealthy interval depending on status
func FireWebhook(data events.SNSEvent) error {
	// Build fields
	fields := []models.DiscordWebhookField{}
	var msg SNSMessage
	err := json.Unmarshal([]byte(data.Records[0].SNS.Message), &msg)
	if err != nil {
		log.Error("Error unmarshalling webhook body", "err", err)
		return err
	}
	fields = append(fields, models.DiscordWebhookField{
		Name:  "Alarm Name",
		Value: msg.AlarmName,
	})
	fields = append(fields, models.DiscordWebhookField{
		Name:  "Region",
		Value: msg.Region,
	})
	fields = append(fields, models.DiscordWebhookField{
		Name:  "New State Reason",
		Value: msg.NewStateReason,
	})
	fields = append(fields, models.DiscordWebhookField{
		Name:  "Alarm Arn",
		Value: msg.AlarmArn,
	})
	// Build webhook body
	body := models.DiscordWebhookBody{
		Embeds: []models.DiscordWebhookEmbed{
			{
				Title:  data.Records[0].SNS.Subject,
				Color:  15548997,
				Fields: fields,
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
	if len(event.Records) == 0 {
		log.Error("No records found in event")
		return "", errors.New("No records found in event")
	}
	// Marshal
	err := FireWebhook(event)
	return "", err
}

func main() {
	lambda.Start(HandleRequest)
}
