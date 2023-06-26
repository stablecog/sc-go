package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
)

func (c *DiscordCommandWrapper) NewAuthenticateCommand() *DiscordCommand {
	return &DiscordCommand{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "authenticate",
			Description: "Connect your Discord Account to Stablecog.",
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				return
			}
			_, err := c.Repo.GetUserByDiscordID(i.Member.User.ID)
			if err != nil && !ent.IsNotFound(err) {
				log.Errorf("Failed to get user by discord ID %v", err)
				responses.UnknownErrorPrivateInteractionResponse(s, i)
				return
			}
			if err != nil && ent.IsNotFound(err) {
				urlComponent, err := responses.URLComponent("Authenticate", "https://stablecog.com")
				if err != nil {
					log.Errorf("Failed to create URL component %v", err)
					responses.UnknownErrorPrivateInteractionResponse(s, i)
					return
				}
				responses.PrivateInteractionResponseWithComponents(s, i, "⚠️ Action Required", "You must authenticate your Discord account with Stablecog before you can use this command.\n\nClick the button below to connect your Discord account to Stablecog.",
					[]discordgo.MessageComponent{
						urlComponent,
					})
				return
			}

			// User is already authenticated
			responses.PrivateInteractionResponse(s, i, "👍", "Your Discord account is already authenticated with Stablecog.")
		},
	}
}
