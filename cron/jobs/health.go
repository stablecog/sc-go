package jobs

import (
	"time"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/upscalemodel"
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

	workerHealthStatus := shared.HEALTHY

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

	lastGenerationTime := time.Now().Add(-24 * time.Hour)
	lastSuccessfulGenerationTime := time.Now().Add(-24 * time.Hour)

	if len(generations) > 0 {
		lastGenerationTime = generations[0].CreatedAt
	}
	if len(successfulGenerations) > 0 {
		lastSuccessfulGenerationTime = successfulGenerations[0].CreatedAt
	}

	// Last successful generation is too old, do a test generation
	var durationMinutes float64 = 3
	if time.Now().Sub(lastSuccessfulGenerationTime).Minutes() > durationMinutes {
		// Assume unhealthy for now
		workerHealthStatus = shared.UNHEALTHY

		// Try to create a test generation
		log.Infof(fmt.Sprintf("%f minutes since last successful generation", durationMinutes))
		err := CreateTestGeneration(log, apiKey)
		if err != nil {
			log.Infof("SC Worker test generation failed -> Assuming unhealthy")
		} else {
			workerHealthStatus = shared.HEALTHY
		}
	}

	log.Infof("Done checking health in %dms", time.Now().Sub(start).Milliseconds())

	// Write health status to redis
	errRedis := j.Redis.SetWorkerHealth(workerHealthStatus)

	if errRedis != nil {
		log.Infof("🔴 Couldn't write SC Worker health status to Redis: %v", errRedis)
	} else {
		log.Infof("🔴 Wrote SC Worker health status to Redis: %d", workerHealthStatus)
	}

	shouldActivateRunpodServerless := false
	activatedRunpodServerless := false

	if workerHealthStatus != shared.HEALTHY {
		shouldActivateRunpodServerless = true
		activated, err := j.ActiveRunpodServerless(log)
		if err != nil {
			log.Errorf("🏃‍♂️‍➡️📦 🔴 Couldn't activate Runpod serverless: %v", err)
		} else {
			log.Infof("🏃‍♂️‍➡️📦 🟢 Activated Runpod serverless: %v", activated)
		}
		activatedRunpodServerless = activated
	}

	return j.Discord.SendDiscordNotificationIfNeeded(
		workerHealthStatus,
		generations,
		lastGenerationTime,
		lastSuccessfulGenerationTime,
		shouldActivateRunpodServerless,
		activatedRunpodServerless,
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

func (j *JobRunner) ActiveRunpodServerless(log Logger) (bool, error) {
	errGen := j.Repo.DB.GenerationModel.Update().Where(
		generationmodel.RunpodEndpointNotNil(),
		generationmodel.RunpodActive(false),
	).SetRunpodActive(true).Exec(j.Ctx)

	if errGen != nil {
		return false, errGen
	}

	errUps := j.Repo.DB.UpscaleModel.Update().Where(
		upscalemodel.RunpodEndpointNotNil(),
		upscalemodel.RunpodActive(false),
	).SetRunpodActive(true).Exec(j.Ctx)

	if errUps != nil {
		return false, errUps
	}

	return true, nil
}

func CreateTestGeneration(log Logger, apiKey string) error {
	log.Infof("Creating test generation to check SC Worker health...")

	if apiKey == "" {
		log.Errorf("SC Worker tester API key not found")
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
