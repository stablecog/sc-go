package jobs

import (
	"sync"
	"time"

	"github.com/stablecog/sc-go/utils/color"
	"k8s.io/klog/v2"
)

const redisStatsPrefix = "stats"

func (j *JobRunner) GetAndSetStats() error {
	start := time.Now()
	klog.Infof("Getting stats...")

	// Stats methods, they map to SQL functions and redis keys
	stats := []string{
		"generation_count",
		"upscale_count",
	}

	var wg sync.WaitGroup
	errors := []error{}
	wg.Add(len(stats))
	for _, value := range stats {
		go func(value string) {
			defer wg.Done()
			err := j.GetAndSetStatFromPostgresToRedis(
				value,
			)
			if err != nil {
				errors = append(errors, err)
			}
		}(value)
	}
	wg.Wait()

	if len(errors) > 0 {
		klog.Errorf("Error getting stats: %v", errors[0])
		return errors[0]

	}

	end := time.Now()
	klog.Infof("Got stats in: %s", color.Green(end.Sub(start).Milliseconds(), "ms"))
	return nil
}

func (j *JobRunner) GetAndSetStatFromPostgresToRedis(
	statsName string,
) error {
	return nil
}
