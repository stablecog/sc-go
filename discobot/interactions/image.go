package interactions

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/discobot/components"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/scworker"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (c *DiscordInteractionWrapper) NewImageCommand() *DiscordInteraction {
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
				numOutputs := 4

				for _, option := range options {
					switch option.Name {
					case "prompt":
						prompt = option.StringValue()
					case "image-count":
						numOutputs = int(option.IntValue())
					}
				}

				// Create context
				ctx := context.Background()
				res, err := scworker.CreateGeneration(
					ctx,
					shared.OperationSourceTypeDiscord,
					nil,
					c.SafetyChecker,
					c.Repo,
					c.Redis,
					c.SMap,
					c.QThrottler,
					u,
					true,
					requests.CreateGenerationRequest{
						Prompt:     prompt,
						NumOutputs: utils.ToPtr[int32](int32(numOutputs)),
					},
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
			shared.OperationSourceTypeDiscord,
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
