package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

const MaxWebhookRetries = 5

type QueueProcessor struct {
	Client *http.Client
}

func NewQueueProcessor() *QueueProcessor {
	return &QueueProcessor{
		Client: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

func (p *QueueProcessor) IssueSCWebhook(data requests.CogWebhookMessage, retries int) int {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return 0
	}

	// Create a new request
	req, err := http.NewRequest("POST", data.Input.WebhookPrivateUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0
	}

	// Set the content type and signature header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("signature", utils.GetEnv().ScWorkerWebhookSecret)

	// Perform the request
	resp, err := p.Client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return 0
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != 200 && resp.StatusCode != 400 && resp.StatusCode != 401 && retries < MaxWebhookRetries {
		fmt.Printf("Webhook failed with status code %d\n", resp.StatusCode)
		sleepTime := 0.15 * float64(retries)
		fmt.Printf("Sleeping %f seconds before retrying webhook\n", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Second)
		return p.IssueSCWebhook(data, retries+1)
	}

	return resp.StatusCode
}
