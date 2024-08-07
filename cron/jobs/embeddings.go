package jobs

import "encoding/json"

const GET_EMBEDDINGS_JOB_NAME = "GET_EMBEDDINGS_JOB"

func (j *JobRunner) GetEmbeddingsAndUpdateQdrant(log Logger) error {
	log.Infof("Running %s...", GET_EMBEDDINGS_JOB_NAME)
	outputs, err := j.Repo.GetOutputsWithNoEmbedding()
	if err != nil {
		log.Errorf("Error getting outputs with no embeddings: %v", err)
		return err
	}

	if len(outputs) > 0 {
		// Convert the first output to JSON
		jsonData, err := json.MarshalIndent(outputs[0], "", "    ")
		if err != nil {
			log.Errorf("Error marshaling output to JSON: %v", err)
			return err
		}

		// Print the JSON-formatted output
		log.Infof("Found outputs with no embeddings: %s", string(jsonData))
	} else {
		log.Infof("No outputs found with no embeddings")
	}

	return nil
}
