package jobs

import (
	"fmt"
	"sync"
	"time"

	"github.com/stablecog/go-apps/utils/color"
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
	rKey := fmt.Sprintf("%s:%s", redisStatsPrefix, statsName)
	res, err := j.Db.QueryContext(j.Ctx, fmt.Sprintf("select %s()", statsName))
	if err != nil {
		return err
	}
	defer res.Close()
	res.Next()
	var data int64
	err = res.Scan(&data)
	if err != nil {
		return err
	}

	errSet := j.Redis.Set(j.Ctx, rKey, data, 0).Err()
	if errSet != nil {
		klog.Errorf("Redis - Error setting '%s': %v", rKey, err)
		return errSet
	}
	klog.Infof("Redis - Set '%s' to '%d' in Redis", rKey, data)
	return nil
}
