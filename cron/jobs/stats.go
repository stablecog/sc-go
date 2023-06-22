package jobs

import (
	"sync"
	"time"
)

func (j *JobRunner) GetUpscaleOutputCount() (int, error) {
	return j.Repo.DB.UpscaleOutput.Query().Count(j.Ctx)
}

func (j *JobRunner) GetGenerationOutputCount() (int, error) {
	return j.Repo.DB.GenerationOutput.Query().Count(j.Ctx)
}

func (j *JobRunner) GetVoiceoverOutputCount() (int, error) {
	return j.Repo.DB.VoiceoverOutput.Query().Count(j.Ctx)
}

func (j *JobRunner) GetAndSetStats(log Logger) error {
	start := time.Now()
	log.Infof("Getting stats...")

	results := make(chan map[string]int, 3)
	errors := make(chan error, 3)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, err := j.GetUpscaleOutputCount()
		if err != nil {
			errors <- err
			return
		}
		results <- map[string]int{
			"upscale_output_count": count,
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, err := j.GetGenerationOutputCount()
		if err != nil {
			errors <- err
			return
		}
		results <- map[string]int{
			"generation_output_count": count,
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, err := j.GetVoiceoverOutputCount()
		if err != nil {
			errors <- err
			return
		}
		results <- map[string]int{
			"voiceover_output_count": count,
		}
	}()

	// Wait all jobs and close channels
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	for err := range errors {
		if err != nil {
			return err
		}
	}

	var generationOutputCount, upscaleOutputCount, voiceoverOutputCount int
	for result := range results {
		resStat, ok := result["generation_output_count"]
		if ok {
			generationOutputCount = resStat
		}
		resStat, ok = result["upscale_output_count"]
		if ok {
			upscaleOutputCount = resStat
		}
		resStat, ok = result["voiceover_output_count"]
		if ok {
			voiceoverOutputCount = resStat
		}
	}

	err := j.Redis.SetOutputCount(generationOutputCount, upscaleOutputCount, voiceoverOutputCount)
	if err != nil {
		return err
	}

	end := time.Now()
	log.Infof("--- upscales %d", upscaleOutputCount)
	log.Infof("--- generations %d", generationOutputCount)
	log.Infof("--- voiceovers %d", voiceoverOutputCount)
	log.Infof("--- Got stats in %dms", end.Sub(start).Milliseconds())
	return nil
}
