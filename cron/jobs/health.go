package jobs

import (
	"time"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Considered failed if len(failures)/len(generations) > maxGenerationFailWithoutNSFWRate
const maxGenerationFailWithoutNSFWRate = 0.5

// Get this number of generations on each check, sorted by created_at DESC
const generationCountToCheck = 10
const successfulGenerationCountToCheck = 1

const HEALTH_JOB_NAME = "HEALTH_JOB"

// CheckHealth cron job
func (j *JobRunner) CheckSCWorkerHealth(log Logger) error {
	start := time.Now()
	log.Infof("Checking health...")
	apiKey := utils.GetEnv().ScWorkerTesterApiKey

	workerHealthStatus := discord.HEALTHY

	generations, err := j.Repo.GetGenerations(generationCountToCheck)
	if err != nil || len(generations) == 0 {
		log.Errorf("Couldn't get generations %v", err)
		return err
	}

	successfulGenerations, err := j.Repo.GetSuccessfulGenerations(successfulGenerationCountToCheck)
	if err != nil || len(generations) == 0 {
		log.Errorf("Couldn't get successful generations %v", err)
		return err
	}

	nsfwGenerations := 0
	failedGenerations := 0
	lastGenerationTime := generations[0].CreatedAt
	lastSuccessfulGenerationTime := successfulGenerations[0].CreatedAt

	// Count the number of failed generations distinguishing between NSFW and other failures
	for _, g := range generations {
		if g.Status == generation.StatusFailed && g.FailureReason != nil && *g.FailureReason == shared.NSFW_ERROR {
			nsfwGenerations++
		} else if g.Status == generation.StatusFailed {
			failedGenerations++
		}
	}

	log.Infof("Generation fail rate NSFW %d/%d", nsfwGenerations, len(generations))
	log.Infof("Generation fail rate other %d/%d", failedGenerations, len(generations))

	// Figure out if we're healthy
	failRate := float64(failedGenerations) / float64(len(generations))

	// Fail rate is too high, fail
	if failRate > maxGenerationFailWithoutNSFWRate {
		workerHealthStatus = discord.UNHEALTHY
	}

	// Last successful generation is too old, do a test generation
	var durationMinutes float64 = 5
	if time.Now().Sub(lastSuccessfulGenerationTime).Minutes() > durationMinutes {
		log.Infof(fmt.Sprintf("%f minutes since last successful generation", durationMinutes))
		if apiKey == "" {
			log.Infof("SC Worker tester API key not found -> Assuming unhealthy")
			workerHealthStatus = discord.UNHEALTHY
		} else {
			err := CreateTestGeneration(log, apiKey)
			if err != nil {
				log.Infof("SC Worker test generation failed -> Assuming unhealthy")
				workerHealthStatus = discord.UNHEALTHY
			}
		}
	}

	log.Infof("Done checking health in %dms", time.Now().Sub(start).Milliseconds())

	// Write health status to redis
	errRedis := j.Redis.Client.Set(j.Redis.Ctx, shared.REDIS_SC_WORKER_HEALTH_KEY, int(workerHealthStatus), 0).Err()
	if errRedis != nil {
		log.Infof("Couldn't write SC Worker health status to Redis: %v", errRedis)
	} else {
		log.Infof("Wrote SC Worker health status to Redis: %d", workerHealthStatus)
	}

	return j.Discord.SendDiscordNotificationIfNeeded(
		workerHealthStatus,
		generations,
		lastGenerationTime,
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

func CreateTestGeneration(log Logger, apiKey string) error {
	log.Infof("Creating test generation to check SC Worker health...")

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
		log.Errorf("SC Worker test generation: Couldn't marshal json %v", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("SC Worker test generation: Couldn't create request %v", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("SC Worker test generation: Couldn't send request %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("SC Worker test generation: Couldn't read response body %v", err)
		return err
	}

	var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		log.Errorf("SC Worker test generation: Couldn't unmarshal response body %v", err)
		return err
	}

	if len(responseBody.Outputs) == 0 {
		log.Errorf("SC Worker test generation: No outputs in response")
		return fmt.Errorf("SC Worker test generation: No outputs in response")
	}

	log.Infof("SC Worker test generation url: %s", responseBody.Outputs[0].ImageURL)

	return nil
}
