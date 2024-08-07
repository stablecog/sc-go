package jobs

const GET_EMBEDDINGS_JOB_NAME = "GET_EMBEDDINGS_JOB"

func (j *JobRunner) GetEmbeddingsAndUpdateQdrant(log Logger) error {
	log.Infof("Running %s...", GET_EMBEDDINGS_JOB_NAME)
	outputs, err := j.Repo.GetOutputsWithNoEmbedding()
	if err != nil {
		log.Errorf("Error getting outputs with no embeddings: %v", err)
		return err
	}
	log.Infof("Found outputs with no embeddings", outputs)
	return nil
}
