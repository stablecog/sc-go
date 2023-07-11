package interactions

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/discobot/aspectratio"
	"github.com/stablecog/sc-go/discobot/components"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/scworker"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (c *DiscordInteractionWrapper) NewImageCommand() *DiscordInteraction {
	// Build model choices
	// Ensure default is at the top
	defaultModel := shared.GetCache().GetDefaultGenerationModel()
	modelChoices := []*discordgo.ApplicationCommandOptionChoice{
		{
			Name:  fmt.Sprintf("%s (default)", defaultModel.NameInWorker),
			Value: defaultModel.ID.String(),
		},
	}
	for _, model := range shared.GetCache().GenerateModels {
		if model.ID == defaultModel.ID {
			continue
		}
		if model.IsActive && !model.IsHidden {
			modelChoices = append(modelChoices, &discordgo.ApplicationCommandOptionChoice{
				Name:  model.NameInWorker,
				Value: model.ID.String(),
			})
		}
	}

	// Build aspect ratio choices
	aspectRatioChoices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, ratio := range aspectratio.AvailableRatios {
		aspectRatioChoices = append(aspectRatioChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  ratio.String(),
			Value: ratio,
		})
	}

	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "image",
			Description: "Create an image with Stablecog.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: "The prompt for the generation.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "image-count",
					Description: "The number of images to generate.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "model",
					Description: "The model for the generation.",
					Required:    false,
					Choices:     modelChoices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "aspect-ratio",
					Description: "The aspect ratio for the generation.",
					Required:    false,
					Choices:     aspectRatioChoices,
				},
			},
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if u := c.Disco.CheckAuthorization(s, i); u != nil {
				// Always create initial message
				responses.InitialLoadingResponse(s, i, responses.PUBLIC)

				// Access options in the order provided by the user.
				options := i.ApplicationCommandData().Options

				// Parse options
				var prompt string
				var modelId *uuid.UUID
				var aspectRatio *aspectratio.AspectRatio
				numOutputs := 4

				for _, option := range options {
					switch option.Name {
					case "prompt":
						prompt = option.StringValue()
					case "image-count":
						numOutputs = int(option.IntValue())
					case "model":
						modelId = utils.ToPtr[uuid.UUID](uuid.MustParse(option.StringValue()))
					case "aspect-ratio":
						aspectRatio = utils.ToPtr(aspectratio.AspectRatio(option.IntValue()))
					}
				}

				if modelId == nil {
					modelId = utils.ToPtr(shared.GetCache().GetDefaultGenerationModel().ID)
				}

				if aspectRatio == nil {
					aspectRatio = utils.ToPtr(aspectratio.DefaultAspectRatio)
				}

				// Validate req/apply defaults
				req := requests.CreateGenerationRequest{
					Prompt:     prompt,
					ModelId:    modelId,
					NumOutputs: utils.ToPtr[int32](int32(numOutputs)),
				}
				if aspectRatio != nil {
					width, height := aspectRatio.GetWidthHeightForModel(*modelId)
					req.Width = utils.ToPtr[int32](width)
					req.Height = utils.ToPtr[int32](height)
				}
				err := req.Validate(true)
				if err != nil {
					responses.ErrorResponseEditValidation(s, i, err.Error())
					return
				}

				// Create context
				ctx := context.Background()
				res, err := scworker.CreateGeneration(
					ctx,
					enttypes.SourceTypeDiscord,
					nil,
					c.SafetyChecker,
					c.Repo,
					c.Redis,
					c.SMap,
					c.QThrottler,
					u,
					true,
					req,
				)
				if err != nil {
					log.Errorf("Error creating generation: %v", err)
					responses.ErrorResponseEdit(s, i)
					return
				}

				var imageUrls []string
				var actionRowOne []*components.SCDiscordComponent
				for i, output := range res.Outputs {
					if output.ImageURL != nil {
						imageUrls = append(imageUrls, *output.ImageURL)
						actionRowOne = append(actionRowOne, components.NewButton(fmt.Sprintf("Upscale #%d", i+1), fmt.Sprintf("upscale:%s:number:%d", output.ID.String(), i+1), "✨"))
					}
				}

				// Send the image
				_, err = responses.InteractionEdit(s, i, &responses.InteractionResponseOptions{
					Content:      utils.ToPtr(fmt.Sprintf("<@%s> **%s**", i.Member.User.ID, prompt)),
					ImageURLs:    imageUrls,
					Embeds:       nil,
					ActionRowOne: actionRowOne,
				},
				)
				if err != nil {
					log.Error(err)
					responses.ErrorResponseEdit(s, i)
				}
			}
		},
	}
}

// Handle upscaling
func (c *DiscordInteractionWrapper) HandleUpscale(s *discordgo.Session, i *discordgo.InteractionCreate, outputId uuid.UUID, number int) {
	if u := c.Disco.CheckAuthorization(s, i); u != nil {
		// Disable the button
		// if len(i.Message.Components) < 1 {
		// 	log.Errorf("Error getting action row")
		// 	responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		// 	return
		// }
		// actionRowRaw, err := i.Message.Components[0].MarshalJSON()
		// if err != nil {
		// 	log.Errorf("Error getting action row: %v", err)
		// 	responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		// 	return
		// }

		// // Parse as SCDiscordComponent
		// var actionRow components.SCDiscordActionRow
		// err = json.Unmarshal(actionRowRaw, &actionRow)
		// if err != nil {
		// 	log.Errorf("Error getting action row: %v", err)
		// 	responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		// 	return
		// }

		// // Get button from action row
		// if len(actionRow.Components) < number {
		// 	log.Errorf("Error getting button")
		// 	responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		// 	return
		// }
		// actionRow.Components[number-1].Disabled = utils.ToPtr(true)
		// marshalled, err := actionRow.AsMessageComponent()
		// if err != nil {
		// 	log.Errorf("Error getting action row: %v", err)
		// 	responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		// 	return
		// }

		// err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseUpdateMessage,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Components: []discordgo.MessageComponent{marshalled},
		// 	},
		// })

		// See if the output is already upscaled, send private response to avoid pollution
		existingOutput, err := c.Repo.GetPublicGenerationOutput(outputId)
		if err != nil {
			log.Errorf("Error getting output: %v", err)
			responses.ErrorResponseInitial(s, i, responses.PRIVATE)
			return
		}
		if existingOutput.UpscaledImagePath != nil {
			// Send the image
			err = responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
				Content: utils.ToPtr(fmt.Sprintf("<@%s> ✨ Image has already been upscaled #%d ✅ \n%s", i.Member.User.ID, number, utils.GetURLFromImagePath(*existingOutput.UpscaledImagePath))),
				Embeds:  nil,
				Privacy: responses.PRIVATE,
			})
			return
		}
		// Always create initial message
		responses.InitialLoadingResponse(s, i, responses.PUBLIC)

		// Create context
		ctx := context.Background()
		res, err := scworker.CreateUpscale(
			ctx,
			enttypes.SourceTypeDiscord,
			nil,
			c.Repo,
			c.Redis,
			c.SMap,
			c.QThrottler,
			u,
			requests.CreateUpscaleRequest{
				Input: outputId.String(),
			},
		)
		if err != nil {
			log.Errorf("Error creating upscale: %v", err)
			responses.ErrorResponseEdit(s, i)
			return
		}

		var upscaledImageUrl string
		for _, output := range res.Outputs {
			if output.UpscaledImageURL != nil {
				upscaledImageUrl = *output.UpscaledImageURL
			}
		}
		if upscaledImageUrl == "" {
			log.Errorf("Error getting upscaled image url")
			responses.ErrorResponseEdit(s, i)
			return
		}

		// Send the image
		_, err = responses.InteractionEdit(s, i, &responses.InteractionResponseOptions{
			Content: utils.ToPtr(fmt.Sprintf("<@%s> ✨ Upscaled #%d ✅ \n%s", i.Member.User.ID, number, upscaledImageUrl)),
			Embeds:  nil,
		},
		)
		if err != nil {
			log.Error(err)
			responses.ErrorResponseEdit(s, i)
		}
	}
}
