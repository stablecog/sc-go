package jobs

import (
	"context"
	"time"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sethvargo/go-retry"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Considered failed if len(failures)/len(generations) > maxGenerationFailWithoutNSFWRate
const maxGenerationFailWithoutNSFWRate = 0.5

// Get this number of generations on each check, sorted by created_at DESC
const generationCountToCheck = 10
const successfulGenerationCountToCheck = 1

// Per-attempt HTTP timeout for the test generation request. A normal successful
// generation takes ~20s, so the timeout has to sit comfortably above that — too
// tight and we'd time out healthy slow-tail requests and flag the worker
// unhealthy for no reason.
const testGenerationHTTPTimeout = 30 * time.Second

// Wall-clock budget for the entire health check. Must accommodate the worst-case
// retry loop: exponentialRetryCount attempts × testGenerationHTTPTimeout, plus
// the exponential backoff sleeps between them (1+2+4 = 7s with the current
// settings). 4 × 30 + 7 = 127s, so 150s leaves a small safety margin.
//
// SingletonMode on the cron schedule means a slow run that consumes most of this
// budget will simply skip the intervening 60s firings — exactly what we want.
// What this bound enforces is: a single run can't drift so far that its
// start-of-run snapshot becomes meaningless by the time the embed is posted.
const healthCheckBudget = 150 * time.Second

const HEALTH_JOB_NAME = "HEALTH_JOB"

// CheckHealth cron job
func (j *JobRunner) CheckSCWorkerHealth(log Logger) error {
	// asOf is sampled once at the start of the run and is the reference time used
	// for every relative-time display + the embed footer. If we used time.Now()
	// at send time instead, a slow run (e.g. one stuck in test-gen retries) would
	// render its old snapshot against a fresh "now" and produce timestamps that
	// look impossible compared to neighbouring fast runs ("20m ago" sandwiched
	// between two "Just now" messages within a single minute).
	asOf := time.Now()
	ctx, cancel := context.WithTimeout(j.Ctx, healthCheckBudget)
	defer cancel()

	log.Infof("Checking health...")
	apiKey := utils.GetEnv().ScWorkerTesterApiKey

	workerHealthStatus := shared.HEALTHY

	generations, err := j.Repo.GetGenerations(generationCountToCheck)
	if err != nil || len(generations) == 0 {
		log.Errorf("Couldn't get generations %v", err)
		return err
	}

	successfulGenerations, err := j.Repo.GetSuccessfulGenerations(successfulGenerationCountToCheck)
	if err != nil {
		log.Errorf("Couldn't get successful generations %v", err)
		return err
	}

	lastGenerationTime := asOf.Add(-24 * time.Hour)
	lastSuccessfulGenerationTime := asOf.Add(-24 * time.Hour)

	if len(generations) > 0 {
		lastGenerationTime = generations[0].CreatedAt
	}
	if len(successfulGenerations) > 0 {
		lastSuccessfulGenerationTime = successfulGenerations[0].CreatedAt
	}

	// Last successful generation is too old, do a test generation
	const durationMinutes float64 = 3
	const exponentialRetryCount uint64 = 4
	const exponentialBaseDelay = 1 * time.Second

	if asOf.Sub(lastSuccessfulGenerationTime).Minutes() > durationMinutes {
		log.Infof(fmt.Sprintf("%d minutes since last successful generation.", int(durationMinutes)))
		b := retry.WithMaxRetries(exponentialRetryCount, retry.NewExponential(exponentialBaseDelay))
		err := retry.Do(ctx, b, func(ctx context.Context) error {
			err := CreateTestGeneration(ctx, log, apiKey)
			if err != nil {
				log.Errorf("🧪 🔴 SC Worker test generation failed: %v", err)
				return retry.RetryableError(err)
			}
			return nil
		})
		if err != nil {
			log.Infof("SC Worker test generation failed -> Assuming unhealthy")
			workerHealthStatus = shared.UNHEALTHY
		}
	}

	log.Infof("Done checking health in %dms", time.Since(asOf).Milliseconds())

	// Write health status to redis
	errRedis := j.Redis.SetWorkerHealth(workerHealthStatus)

	if errRedis != nil {
		log.Infof("🔴 Couldn't write SC Worker health status to Redis: %v", errRedis)
	} else {
		log.Infof("🟢 Wrote SC Worker health status to Redis: %d", workerHealthStatus)
	}

	isRunpodServerlessActive, runpodServerlessErr := j.Repo.IsRunpodServerlessActive()

	if runpodServerlessErr != nil {
		log.Errorf("🏃‍♂️‍➡️📦 🔴 Couldn't check if Runpod serverless is active: %v", runpodServerlessErr)
	}

	if isRunpodServerlessActive {
		log.Infof("🏃‍♂️‍➡️📦 🟢 Runpod serverless is active")
	}

	if workerHealthStatus != shared.HEALTHY && !isRunpodServerlessActive {
		err := j.Repo.EnableRunpodServerless()
		if err != nil {
			log.Errorf("🏃‍♂️‍➡️📦 🔴 Couldn't activate Runpod serverless: %v", err)
			runpodServerlessErr = err
		} else {
			log.Infof("🏃‍♂️‍➡️📦 🟢 Activated Runpod serverless")
			isRunpodServerlessActive = true
		}
	}

	return j.Discord.SendDiscordNotificationIfNeeded(
		workerHealthStatus,
		generations,
		lastGenerationTime,
		lastSuccessfulGenerationTime,
		isRunpodServerlessActive,
		runpodServerlessErr,
		asOf,
	)
}

type RequestBody struct {
	Prompt     string `json:"prompt"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	NumOutputs int    `json:"num_outputs"`
}

type ResponseBody struct {
	Outputs []struct {
		ID       string `json:"id"`
		URL      string `json:"url"`
		ImageURL string `json:"image_url"`
	} `json:"outputs"`
}

func CreateTestGeneration(ctx context.Context, log Logger, apiKey string) error {
	log.Infof("🧪 Creating test generation to check SC Worker health...")

	if apiKey == "" {
		log.Errorf("🧪 🔴 SC Worker tester API key not found")
		return fmt.Errorf("SC Worker tester API key not found")
	}

	url := "https://api.stablecog.com/v1/image/generation/create"
	prompt := "Mavi renkli bir bina"
	width := 1024
	height := 1024
	numOutputs := 1

	requestBody := RequestBody{
		Prompt:     prompt,
		Width:      width,
		Height:     height,
		NumOutputs: numOutputs,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't marshal json %v", err)
		debugString := fmt.Sprintf("%+v", requestBody)
		log.Errorf("🧪 🔴 Request body: %s", debugString)
		return err
	}

	// Pin a per-attempt deadline so a stalled API call can't drag the cron job
	// out beyond healthCheckBudget. The outer ctx is still honored — whichever
	// expires first wins.
	reqCtx, cancel := context.WithTimeout(ctx, testGenerationHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't create request %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: testGenerationHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't send request %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't read response body %v", err)
		return err
	}

	var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't unmarshal response body %v", err)
		return err
	}

	if len(responseBody.Outputs) == 0 {
		log.Errorf("🧪 🔴 No outputs in response")
		return fmt.Errorf("SC Worker test generation: No outputs in response")
	}

	log.Infof("🧪 🟢 SC Worker test generation created: %s", responseBody.Outputs[0].ImageURL)

	return nil
}
