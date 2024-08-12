package jobs

import (
	"time"

	"github.com/stablecog/sc-go/utils"
)

const NSFW_CHECK_JOB_NAME = "NSFW_CHECK_JOB"
const NSFW_CHECK_OUTPUTS_LIMIT = 10

func (j *JobRunner) HandleOutputsWithNoNsfwCheck(log Logger) error {
	log.Infof("Running job...")
	s := time.Now()

	outputs, err := j.Repo.GetOutputsWithNoNsfwCheck(NSFW_CHECK_OUTPUTS_LIMIT)
	if err != nil {
		log.Errorf("Error getting outputs with no NSFW check: %v", err)
		return err
	}

	m := time.Since(s)

	if len(outputs) < 1 {
		log.Infof("No outputs found with no NSFW check: %dms", m.Milliseconds())
		return nil
	}

	log.Infof("Found %d outputs with no NSFW check: %dms", len(outputs), m.Milliseconds())
	log.Infof("Getting NSFW scores for outputs...")

	var imageUrls []string

	for _, output := range outputs {
		imageUrls = append(imageUrls, utils.GetEnv().GetURLFromImagePath(output.ImagePath))
	}

	nsfwScores, err := j.CLIP.GetNsfwScores(imageUrls)

	if err != nil {
		log.Errorf("Error getting NSFW scores: %v", err)
		return err
	}

	log.Infof("Got NSFW scores for %d output(s)", len(nsfwScores))

	e := time.Since(s)

	log.Infof("✅ Job completed | %d item(s) | %dms", len(outputs), e.Milliseconds())

	return nil
}
