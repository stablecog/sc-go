package jobs

import "time"

const EMBEDDINGS_JOB_NAME = "EMBEDDINGS_JOB"

func (j *JobRunner) HandleOutputsWithNoEmbedding(log Logger) error {
	log.Infof("Running job...")
	s := time.Now()

	outputs, err := j.Repo.GetOutputsWithNoEmbedding()
	if err != nil {
		log.Errorf("Error getting outputs with no embeddings: %v", err)
		return err
	}

	e := time.Since(s)
	if len(outputs) > 0 {
		log.Infof("Found %d outputs with no embeddings: %dms", len(outputs), e.Milliseconds())
	} else {
		log.Infof("No outputs found with no embeddings: %dms", e.Milliseconds())
	}

	return nil
}
