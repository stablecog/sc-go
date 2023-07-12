package interactions

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/discobot/responses"
)

type LoginInteraction struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	InsertedAt  time.Time
}

func (c *DiscordInteractionWrapper) NewAuthenticateCommand() *DiscordInteraction {
	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "authenticate",
			Description: "Connect your Discord Account to Stablecog.",
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
				// User is already authenticated
				responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
					EmbedTitle:   "👍",
					EmbedContent: "Your Discord account is already authenticated with Stablecog.",
				})
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
