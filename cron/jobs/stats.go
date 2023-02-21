package jobs

import (
	"sync"
	"time"

	"github.com/stablecog/sc-go/utils/color"
	"k8s.io/klog/v2"
)

func (j *JobRunner) GetUpscaleOutputCount() (int, error) {
	return j.Repo.DB.UpscaleOutput.Query().Count(j.Ctx)
}

func (j *JobRunner) GetGenerationOutputCount() (int, error) {
	return j.Repo.DB.GenerationOutput.Query().Count(j.Ctx)
}

func (j *JobRunner) GetAndSetStats() error {
	start := time.Now()
	klog.Infof("Getting stats...")

	results := make(chan map[string]int, 2)
	errors := make(chan error, 2)

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
			"upscale_count": count,
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
			"generation_count": count,
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

	var generationCount, upscaleCount int
	for result := range results {
		resStat, ok := result["generation_count"]
		if ok {
			generationCount = resStat
		}
		resStat, ok = result["upscale_count"]
		if ok {
			upscaleCount = resStat
		}
	}

	err := j.Redis.SetGenerateUpscaleCount(generationCount, upscaleCount)
	if err != nil {
		return err
	}

	end := time.Now()
	klog.Infof("--- upscales: %s", color.Green(upscaleCount))
	klog.Infof("--- generations: %s", color.Green(generationCount))
	klog.Infof("--- Got stats in: %s", color.Green(end.Sub(start).Milliseconds(), "ms"))
	return nil
}
