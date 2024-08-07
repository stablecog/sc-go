package jobs

import (
	"time"

	"github.com/stablecog/sc-go/server/clip"
	"github.com/stablecog/sc-go/utils"
)

const EMBEDDINGS_JOB_NAME = "EMBEDDINGS_JOB"

func (j *JobRunner) HandleOutputsWithNoEmbedding(log Logger) error {
	log.Infof("Running job...")
	s := time.Now()

	outputs, err := j.Repo.GetOutputsWithNoEmbedding(10)
	if err != nil {
		log.Errorf("Error getting outputs with no embeddings: %v", err)
		return err
	}

	m := time.Since(s)

	if len(outputs) < 1 {
		log.Infof("No outputs found with no embeddings: %dms", m.Milliseconds())
		return nil
	}

	log.Infof("Found %d outputs with no embeddings: %dms", len(outputs), m.Milliseconds())
	log.Infof("Getting embeddings for outputs...")

	for _, output := range outputs {
		tOutput := time.Now()
		_, err := j.CLIP.GetEmbeddingsV2([]clip.EmbeddingReqObject{
			{
				Image:          utils.GetEnv().GetURLFromImagePath(output.ImagePath),
				CalculateScore: true,
			},
		})
		if err != nil {
			log.Errorf(`Error getting embeddings for "%s": %v`, output.ID.String(), err)
			continue
		}
		mOutput := time.Since(tOutput)
		log.Infof(`Got embeddings for "%s": %dms`, output.ID.String(), mOutput.Milliseconds())
	}

	e := time.Since(s)

	log.Infof("Job complete: %dms", e.Milliseconds())

	return nil
}
