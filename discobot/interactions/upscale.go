package interactions

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/scworker"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (c *DiscordInteractionWrapper) NewUpscaleCommand() *DiscordInteraction {
	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "upscale",
			Description: "Upscale an image with Stablecog.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "image",
					Description: "The image to upscale.",
					Required:    true,
				},
			},
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var discordUserId string
			if i.Member != nil {
				discordUserId = i.Member.User.ID
			} else {
				discordUserId = i.User.ID
			}
			if u := c.Disco.CheckAuthorization(s, i); u != nil {
				// Get attachment
				if i.ApplicationCommandData().Resolved == nil {
					log.Errorf("No resolved data for upscale command: %v", i.ApplicationCommandData())
					responses.ErrorResponseInitial(s, i, responses.PRIVATE)
					return
				}

				// Access options in the order provided by the user.
				options := i.ApplicationCommandData().Options

				// Parse options
				var attachmentId string

				for _, option := range options {
					switch option.Name {
					case "image":
						id, ok := option.Value.(string)
						if !ok {
							log.Errorf("Invalid image attachment for upscale command: %v", i.ApplicationCommandData())
							responses.ErrorResponseInitial(s, i, responses.PRIVATE)
							return
						}
						attachmentId = id
					}
				}

				attachment, ok := i.ApplicationCommandData().Resolved.Attachments[attachmentId]
				if !ok {
					log.Errorf("No image attachment for upscale command: %v", i.ApplicationCommandData())
					responses.ErrorResponseInitial(s, i, responses.PRIVATE)
					return
				}

				if attachment.ContentType != "image/png" && attachment.ContentType != "image/jpeg" && attachment.ContentType != "image/jpg" && attachment.ContentType != "image/webp" {
					responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
						Privacy:      responses.PRIVATE,
						EmbedTitle:   "❌ Attachment type is not supported",
						EmbedContent: "The attachment can be a PNG, JPEG, or WEBP image.",
					})
					return
				}

				if attachment.Width > shared.MAX_UPSCALE_INITIAL_HEIGHT || attachment.Width > shared.MAX_UPSCALE_INITIAL_WIDTH {
					responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
						Privacy:      responses.PRIVATE,
						EmbedTitle:   "❌ Image is too large",
						EmbedContent: fmt.Sprintf("The width can't be higher than %dpx, and the height can't be higher than %dpx.", shared.MAX_UPSCALE_INITIAL_WIDTH, shared.MAX_UPSCALE_INITIAL_HEIGHT),
					})
					return
				}

				// Do the upscale
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
						Input: attachment.URL,
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
					Content: utils.ToPtr(fmt.Sprintf("<@%s> ✨ Upscaled %s \n%s", discordUserId, attachment.Filename, upscaledImageUrl)),
					Embeds:  nil,
				},
				)
				if err != nil {
					log.Error(err)
					responses.ErrorResponseEdit(s, i)
				}
			} else {
				c.LoginInteractionMap.Put(discordUserId, &LoginInteraction{
					Session:     s,
					Interaction: i,
					InsertedAt:  time.Now(),
				})
			}
		},
	}
}
