package interactions

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/discobot/responses"
)

func (c *DiscordInteractionWrapper) NewAuthenticateCommand() *DiscordInteraction {
	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "authenticate",
			Description: "Connect your Discord Account to Stablecog.",
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if u := c.Disco.CheckAuthorization(s, i); u != nil {
				// User is already authenticated
				responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
					EmbedTitle:   "👍",
					EmbedContent: "Your Discord account is already authenticated with Stablecog.",
				})
			}
		},
	}
}
