package jobs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
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

	if len(outputs) < 1 {
		log.Infof("No outputs found with no NSFW check: %dms", time.Since(s).Milliseconds())
		return nil
	}

	log.Infof("Found %d outputs with no NSFW check: %dms", len(outputs), time.Since(s).Milliseconds())
	log.Infof("Getting NSFW scores for outputs...")

	m := time.Now()
	var imageUrls []string

	for _, output := range outputs {
		imageUrls = append(imageUrls, utils.GetEnv().GetURLFromImagePath(output.ImagePath))
	}

	nsfwScores, err := j.CLIP.GetNsfwScores(imageUrls)

	if err != nil {
		log.Errorf("Error getting NSFW scores: %v", err)
		return err
	}

	log.Infof("Got NSFW scores for %d output(s): %dms", len(nsfwScores), time.Since(m).Milliseconds())

	log.Infof("Updating outputs with NSFW scores...")
	m = time.Now()
	r := j.Repo
	for i, output := range outputs {
		if err := r.WithTx(func(tx *ent.Tx) error {
			score := nsfwScores[i]
			db := tx.Client()
			_, goErr := db.GenerationOutput.
				UpdateOneID(output.ID).
				SetCheckedForNsfw(true).
				SetNsfwScore(score).
				Save(r.Ctx)

			if goErr != nil {
				log.Errorf("Error updating output with NSFW score: %s | Error: %v", output.ID.String(), goErr)
				return goErr
			}

			if j.Qdrant == nil {
				log.Infof("Qdrant client not initialized, not adding to qdrant")
				return fmt.Errorf("Qdrant client not initialized")
			}

			payload := map[string]interface{}{
				"nsfw_score": score,
			}
			err = j.Qdrant.SetPayload(payload, []uuid.UUID{output.ID}, false)

			if err != nil {
				log.Errorf("Error updating Qdrant with NSFW score | ID: %s | Err: %v", output.ID.String(), err)
				return err
			}

			return nil
		}); err != nil {
			log.Errorf("Error starting transaction in HandleOutputsWithNoNsfwCheck: %s | Error: %v", output.ID.String(), err)
			continue
		}
	}
	log.Infof("Updated %d output(s) with NSFW scores in Postgres & Qdrant: %dms", len(outputs), time.Since(m).Milliseconds())

	e := time.Since(s)

	log.Infof("✅ Job completed | %d item(s) | %dms", len(outputs), e.Milliseconds())

	return nil
}
