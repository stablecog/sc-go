package interactions

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	srvresponses "github.com/stablecog/sc-go/server/responses"
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

				if attachment.Width*attachment.Height > shared.MAX_UPSCALE_MEGAPIXELS {
					responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
						Privacy:      responses.PRIVATE,
						EmbedTitle:   "❌ Image is too large",
						EmbedContent: fmt.Sprintf("The image can't be larger than %.1f megapixels.", float64(shared.MAX_UPSCALE_MEGAPIXELS/1000000)),
					})
					return
				}

				req := requests.CreateUpscaleRequest{
					Input: attachment.URL,
				}

				credits, err := c.Repo.GetNonExpiredCreditTotalForUser(u.ID, nil)
				if err != nil {
					log.Errorf("Error getting credits for user: %v", err)
					responses.ErrorResponseInitial(s, i, responses.PRIVATE)
					return
				}
				if credits < int(req.Cost()) {
					responses.InitialInteractionResponse(s, i, responses.InsufficientCreditsResponseOptions(req.Cost(), int32(credits)))
					return
				}

				// Do the upscale
				// Always create initial message
				responses.InitialLoadingResponse(s, i, responses.PUBLIC)

				// Create upscale
				res, _, wErr := c.SCWorker.CreateUpscale(
					enttypes.SourceTypeDiscord,
					nil,
					u,
					nil,
					req,
				)
				if wErr != nil {
					if errors.Is(wErr.Err, srvresponses.InsufficientCreditsErr) {
						credits, err := c.Repo.GetNonExpiredCreditTotalForUser(u.ID, nil)
						if err != nil {
							log.Errorf("Error getting credits for user: %v", err)
							responses.ErrorResponseEdit(s, i)
							return
						}
						responses.InteractionEdit(s, i, responses.InsufficientCreditsResponseOptions(req.Cost(), int32(credits)))
						return
					}
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

				err = c.Repo.UpdateLastSeenAt(u.ID)
				if err != nil {
					log.Warn("Error updating last seen at", "err", err, "user", u.ID.String())
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
