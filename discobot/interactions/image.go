package interactions

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/scworker"
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
					Name:        "num-outputs",
					Description: "The number of outputs to generate.",
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
					case "num-outputs":
						numOutputs = int(option.IntValue())
					}
				}

				// Create context
				ctx := context.Background()
				res, err := scworker.CreateGeneration(
					ctx,
					c.Repo,
					c.Redis,
					c.SMap,
					c.QThrottler,
					u,
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
				for _, output := range res.Outputs {
					if output.ImageURL != nil {
						imageUrls = append(imageUrls, *output.ImageURL)
					}
				}

				// Send the image
				_, err = responses.InteractionEdit(s, i, &responses.InteractionResponseOptions{
					Content:    utils.ToPtr(fmt.Sprintf("<@%s> **%s**", i.Member.User.ID, prompt)),
					ImageURLs:  imageUrls,
					Embeds:     nil,
					Components: nil,
				},
				)
				if err != nil {
					log.Error(err)
				}
			}
		},
	}
}
