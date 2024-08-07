package jobs

const GET_EMBEDDINGS_JOB_NAME = "GET_EMBEDDINGS_JOB"

func (j *JobRunner) GetEmbeddingsAndUpdateQdrant(log Logger) error {
	log.Infof("Running %s...", GET_EMBEDDINGS_JOB_NAME)
	j.Repo.GetOutputsWithNoEmbedding()
	return nil
}
