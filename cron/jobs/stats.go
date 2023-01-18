package jobs

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/stablecog/go-apps/utils/color"
	"k8s.io/klog/v2"
)

const redisStatsPrefix = "stats"
const statsTTL = 10 * time.Second // redis expiry

func (j *JobRunner) GetAndSetStats() error {
	start := time.Now()
	klog.Infof("Getting stats...")

	// Map for stats
	stats := map[string]*int64{
		"generation_count": nil,
		"upscale_count":    nil,
	}

	var wg sync.WaitGroup
	errors := []error{}
	wg.Add(len(stats))
	for name, value := range stats {
		go func(name string, value *int64) {
			defer wg.Done()
			err := j.GetAndSetStatFromPostgresToRedis(
				name,
				value,
			)
			if err != nil {
				errors = append(errors, err)
			}
		}(name, value)
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
	statsValue *int64,
) error {
	rKey := fmt.Sprintf("%s:%s", redisStatsPrefix, statsName)
	val := j.Redis.Get(j.Ctx, rKey).Val()
	if val != "" {
		num, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			*statsValue = num
			klog.Infof("Redis - Got '%s' from Redis, skipping Supabase", rKey)
			return nil
		}
	}
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

	statsValue = &data
	errSet := j.Redis.Set(j.Ctx, rKey, data, statsTTL).Err()
	if errSet != nil {
		klog.Errorf("Redis - Error setting '%s': %v", rKey, err)
		return errSet
	}
	klog.Infof("Redis - Set '%s' to '%d' in Redis", rKey, data)
	return nil
}
