package interactions

import (
	"context"

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
			},
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if u := c.Disco.CheckAuthorization(s, i); u != nil {
				// Access options in the order provided by the user.
				options := i.ApplicationCommandData().Options

				// Prompt
				prompt := options[0].StringValue()
				ctx := context.Background()
				// Generate the image
				log.Infof("Generating image for %s", u.ID)
				res, err := scworker.CreateGeneration(
					ctx,
					c.Repo,
					c.Redis,
					c.SMap,
					c.QThrottler,
					u,
					requests.CreateGenerationRequest{
						Prompt:     prompt,
						NumOutputs: utils.ToPtr[int32](1),
					},
				)
				if err != nil {
					responses.PrivateInteractionResponse(s, i, "👍", "Your Discord account is already authenticated with Stablecog.", "")
				}

				// Send the image
				err = responses.PublicImageResponse(s, i, *res.Outputs[0].ImageURL, nil)
				if err != nil {
					log.Error(err)
				}
			}
		},
	}
}
