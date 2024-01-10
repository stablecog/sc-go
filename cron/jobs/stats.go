package jobs

import (
	"fmt"
	"sync"
	"time"

	"github.com/stripe/stripe-go/v74"
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

func (j *JobRunner) GetMRR() (int, error) {
	fmt.Print("***** Getting MRR...")
	var totalMRR int = 0
	params := &stripe.SubscriptionListParams{}
	params.Filters.AddFilter("status", "", "active")
	params.AddExpand("data.default_payment_method")
	i := j.Stripe.Subscriptions.List(params)

	for i.Next() {
		fmt.Print("***** Getting Subscription...")
		s := i.Subscription()
		// Assuming all subscriptions are monthly. Adjust logic for other billing cycles
		for _, item := range s.Items.Data {
			totalMRR += int(item.Price.UnitAmount * item.Quantity)
		}
	}

	if err := i.Err(); err != nil {
		return 0, err
	}

	return totalMRR, nil
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		amount, err := j.GetMRR()
		if err != nil {
			errors <- err
			return
		}
		results <- map[string]int{
			"mrr": amount,
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

	var generationOutputCount, upscaleOutputCount, voiceoverOutputCount, mrr int
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
		resStat, ok = result["mrr"]
		if ok {
			mrr = resStat
		}
	}

	err := j.Redis.SetOutputCount(generationOutputCount, upscaleOutputCount, voiceoverOutputCount, mrr)
	if err != nil {
		return err
	}

	end := time.Now()
	log.Infof("--- upscales %d", upscaleOutputCount)
	log.Infof("--- generations %d", generationOutputCount)
	log.Infof("--- voiceovers %d", voiceoverOutputCount)
	log.Infof("--- mrr %d", mrr)
	log.Infof("--- Got stats in %dms", end.Sub(start).Milliseconds())
	return nil
}
