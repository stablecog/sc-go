package jobs

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/utils"
)

const DATA_DELETE_JOB = "USER_DATA_DELETE_JOB"

func (j *JobRunner) DeleteUserData(log Logger, dryRun bool) error {
	usersBanned, err := j.Repo.GetBannedUsersToDelete()
	if err != nil {
		log.Errorf("Error getting users to delete %v", err)
		return err
	}
	usersNotBanned, err := j.Repo.GetUsersToDelete()
	if err != nil {
		log.Errorf("Error getting users to delete %v", err)
		return err
	}

	// Combine banned and not banned users
	users := make(map[uuid.UUID]*ent.User)
	for _, u := range usersBanned {
		users[u.ID] = u
	}
	for _, u := range usersNotBanned {
		users[u.ID] = u
	}

	if len(users) == 0 {
		log.Infof("No users to delete")
		return nil
	}

	// Get all generation outputs for these users
	grandTotalOutputs := 0
	grandTotalGenerations := 0
	grandTotalPrompts := 0
	grandTotalNegativePrompts := 0
	for _, u := range users {
		log.Infof("Deleting user data %s", u.ID)
		outputs, err := j.Repo.GetUserGenerationOutputs(u.ID)
		if err != nil {
			log.Errorf("Error getting generation outputs for user %s: %v", u.ID, err)
			return err
		}

		log.Infof("Deleting %d outputs for user %s", len(outputs), u.ID)
		grandTotalOutputs += len(outputs)

		// Get list of S3 objects
		var paths []*s3.ObjectIdentifier
		for _, output := range outputs {
			paths = append(paths, &s3.ObjectIdentifier{
				Key: aws.String(output.ImagePath),
			})
			if output.UpscaledImagePath != nil {
				paths = append(paths, &s3.ObjectIdentifier{
					Key: aws.String(*output.UpscaledImagePath),
				})
			}
		}

		// Delete S3 objects
		if !dryRun && len(paths) > 0 {
			o, err := j.S3.DeleteObjects(&s3.DeleteObjectsInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET_NAME")),
				Delete: &s3.Delete{
					Objects: paths,
				},
			})
			if err != nil {
				log.Errorf("Error deleting objects for user %s: %v", u.ID, err)
				return err
			}

			log.Infof("Deleted %d objects for user %s", len(o.Deleted), u.ID)
		} else {
			for _, path := range paths {
				log.Infof("Would delete %s", *path.Key)
			}
		}

		// Delete all uploaded objects
		hashedId := utils.Sha256(u.ID.String())
		out, err := j.S3Img2Img.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
			Prefix: aws.String(fmt.Sprintf("%s/", hashedId)),
		})
		if err != nil {
			log.Errorf("Error listing img2img objects for user %s: %v", u.ID, err)
			return err
		}
		var img2imgPaths []*s3.ObjectIdentifier
		for _, obj := range out.Contents {
			img2imgPaths = append(img2imgPaths, &s3.ObjectIdentifier{
				Key: obj.Key,
			})
		}
		if len(img2imgPaths) > 0 {
			if !dryRun {
				log.Infof("Deleting %d img2img objects for user %s", len(img2imgPaths), u.ID)
				_, err = j.S3Img2Img.DeleteObjects(&s3.DeleteObjectsInput{
					Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
					Delete: &s3.Delete{
						Objects: img2imgPaths,
					},
				})
				if err != nil {
					log.Errorf("Error deleting img2img objects for user %s: %v", u.ID, err)
					return err
				}
			} else {
				for _, path := range img2imgPaths {
					log.Infof("Would delete upload %s", *path.Key)
				}
			}
		}

		// Get generations for these IDs
		generationIds := make([]uuid.UUID, len(outputs))
		outputIds := make([]uuid.UUID, len(outputs))
		for i, output := range outputs {
			generationIds[i] = output.GenerationID
			outputIds[i] = output.ID
		}

		generations, err := j.Repo.GetGenerationsByIDList(generationIds)
		if err != nil {
			log.Errorf("Error getting generations for user %s: %v", u.ID, err)
			return err
		}
		grandTotalGenerations += len(generations)

		var promptIds []uuid.UUID
		var negativePromptIds []uuid.UUID
		for _, g := range generations {
			if g.PromptID != nil {
				promptIds = append(promptIds, *g.PromptID)
			}
			if g.NegativePromptID != nil {
				negativePromptIds = append(negativePromptIds, *g.NegativePromptID)
			}
		}

		// Filter out to unique prompts
		promptsToRemove, err := j.Repo.GetUsersUniquePromptIds(promptIds, u.ID)
		if err != nil {
			log.Errorf("Error getting unique prompts for user %s: %v", u.ID, err)
			return err
		}
		grandTotalPrompts += len(promptsToRemove)
		negativePromptsToRemove, err := j.Repo.GetUsersUniqueNegativePromptIds(negativePromptIds, u.ID)
		if err != nil {
			log.Errorf("Error getting unique negative prompts for user %s: %v", u.ID, err)
			return err
		}
		grandTotalNegativePrompts += len(negativePromptsToRemove)

		if !dryRun {
			if err := j.Repo.WithTx(func(tx *ent.Tx) error {
				// Delete generation outputs
				if _, err := tx.GenerationOutput.Delete().Where(
					generationoutput.IDIn(outputIds...),
				).Exec(j.Ctx); err != nil {
					log.Errorf("Error deleting generation outputs for user %s: %v", u.ID, err)
					return err
				}

				// Delete generations
				if _, err := tx.Generation.Delete().Where(
					generation.IDIn(generationIds...),
				).Exec(j.Ctx); err != nil {
					log.Errorf("Error deleting generations for user %s: %v", u.ID, err)
					return err
				}

				// Delete prompts
				if _, err := tx.Prompt.Delete().Where(
					prompt.IDIn(promptsToRemove...),
				).Exec(j.Ctx); err != nil {
					log.Errorf("Error deleting prompts for user %s: %v", u.ID, err)
					return err
				}

				// Delete negative prompts
				if _, err := tx.NegativePrompt.Delete().Where(
					negativeprompt.IDIn(negativePromptsToRemove...),
				).Exec(j.Ctx); err != nil {
					log.Errorf("Error deleting negative prompts for user %s: %v", u.ID, err)
					return err
				}

				// Delete from qdrant
				if err := j.Qdrant.DeleteAllIDs(outputIds, false); err != nil {
					log.Errorf("Error deleting from qdrant for user %s: %v", u.ID, err)
					return err
				}

				// Set deleted_at on user
				if _, err := tx.User.UpdateOneID(u.ID).SetDataDeletedAt(time.Now()).Save(j.Ctx); err != nil {
					log.Errorf("Error setting deleted_at for user %s: %v", u.ID, err)
					return err
				}

				return nil
			}); err != nil {
				log.Errorf("Error in TX %s: %v", u.ID, err)
				return err
			}
		} else {
			for _, id := range outputIds {
				log.Infof("Would delete output %s", id)
			}
			for _, id := range generationIds {
				log.Infof("Would delete generation %s", id)
			}
			for _, id := range promptsToRemove {
				log.Infof("Would delete prompt %s", id)
			}
			for _, id := range negativePromptsToRemove {
				log.Infof("Would delete negative prompt %s", id)
			}
		}
	}
	log.Infof("Total outputs %d", grandTotalOutputs)
	log.Infof("Total generations %d", grandTotalGenerations)
	log.Infof("Total prompts %d", grandTotalPrompts)
	log.Infof("Total negative prompts %d", grandTotalNegativePrompts)
	log.Infof("Total users %d", len(users))
	return nil
}
