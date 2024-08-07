package jobs

import (
	"fmt"
	"time"

	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/clip"
	"github.com/stablecog/sc-go/utils"
)

const EMBEDDINGS_JOB_NAME = "EMBEDDINGS_JOB"
const OUTPUTS_LIMIT = 10

func (j *JobRunner) HandleOutputsWithNoEmbedding(log Logger) error {
	log.Infof("Running job...")
	s := time.Now()

	outputs, err := j.Repo.GetOutputsWithNoEmbedding(OUTPUTS_LIMIT)
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
		embeddingRes, err := j.CLIP.GetEmbeddings([]clip.EmbeddingReqObject{
			{
				Image:          utils.GetEnv().GetURLFromImagePath(output.ImagePath),
				CalculateScore: true,
			},
		})

		if err != nil {
			log.Errorf(`Error getting embeddings for "%s": %v`, output.ID.String(), err)
			continue
		}

		if len(embeddingRes) < 1 {
			log.Errorf(`No embeddings found for "%s"`, output.ID.String())
			continue
		}

		embedding := embeddingRes[0]

		mOutput := time.Since(tOutput)
		log.Infof(`Got embeddings for "%s": %dms`, output.ID.String(), mOutput.Milliseconds())

		r := j.Repo

		if err := r.WithTx(func(tx *ent.Tx) error {
			db := tx.Client()
			_, goErr := db.GenerationOutput.
				UpdateOneID(output.ID).
				SetHasEmbeddings(true).
				SetAestheticArtifactScore(embedding.AestheticScore.Artifact).
				SetAestheticRatingScore(embedding.AestheticScore.Rating).
				Save(r.Ctx)

			if goErr != nil {
				log.Errorf("Error updating output with embeddings: %s | Error: %v", output.ID.String(), goErr)
				return goErr
			}

			if j.Qdrant == nil {
				log.Infof("Qdrant client not initialized, not adding to qdrant")
				return fmt.Errorf("Qdrant client not initialized")
			}

			// Qdrant update
			generation := output.Edges.Generations
			promptObj := generation.Edges.Prompt
			negativePromptObj := generation.Edges.NegativePrompt

			if generation == nil {
				log.Errorf("Generation object not found for output: %s", output.ID.String())
				return fmt.Errorf("Generation object not found for output")
			}

			if promptObj == nil {
				log.Errorf("Prompt object not found for generation: %s", generation.ID.String())
				return fmt.Errorf("Prompt object not found for generation")
			}

			payload := map[string]interface{}{
				"image_path":               output.ImagePath,
				"gallery_status":           output.GalleryStatus,
				"is_favorited":             output.IsFavorited,
				"created_at":               output.CreatedAt.Unix(),
				"updated_at":               output.UpdatedAt.Unix(),
				"is_public":                output.IsPublic,
				"aesthetic_rating_score":   output.AestheticRatingScore,
				"aesthetic_artifact_score": output.AestheticArtifactScore,
				"was_auto_submitted":       generation.WasAutoSubmitted,
				"guidance_scale":           generation.GuidanceScale,
				"inference_steps":          generation.InferenceSteps,
				"prompt_strength":          generation.PromptStrength,
				"height":                   generation.Height,
				"width":                    generation.Width,
				"model":                    generation.ModelID.String(),
				"scheduler":                generation.SchedulerID.String(),
				"user_id":                  generation.UserID.String(),
				"generation_id":            generation.ID.String(),
				"prompt_id":                promptObj.ID.String(),
				"prompt":                   promptObj.Text,
			}
			if output.UpscaledImagePath != nil {
				payload["upscaled_image_path"] = *output.UpscaledImagePath
			}
			if generation.InitImageURL != nil {
				payload["init_image_url"] = generation.InitImageURL
			}
			if negativePromptObj != nil && negativePromptObj.Text != "" {
				payload["negative_prompt"] = negativePromptObj.Text
			}
			err = j.Qdrant.Upsert(
				output.ID,
				payload,
				embedding.Embedding,
				false,
			)

			if err != nil {
				log.Errorf("Error upserting to Qdrant | ID: %s | Err: %v", output.ID.String(), err)
				return err
			}

			return nil
		}); err != nil {
			log.Errorf("Error starting transaction in HandleOutputsWithNoEmbedding: %s | Error: %v", output.ID.String(), err)
			continue
		}
	}

	e := time.Since(s)

	log.Infof("Job complete | %d item(s) | %dms", len(outputs), e.Milliseconds())

	return nil
}
