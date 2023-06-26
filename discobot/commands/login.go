package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
)

func (c *DiscordCommands) AuthenticateCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "authenticate",
		Description: "Connect your Discord Account to Stablecog.",
	}
}

func (c *DiscordCommands) AuthenticateHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member == nil {
		return
	}
	_, err := c.Repo.GetUserByDiscordID(i.Member.User.ID)
	if err != nil && !ent.IsNotFound(err) {
		log.Errorf("Failed to get user by discord ID %v", err)
		responses.PrivateInteractionResponse(s, i, "Something went wrong. Please try again later.")
		return
	}
	if err != nil && ent.IsNotFound(err) {
		urlComponent, err := responses.URLComponent("Authenticate", "https://stablecog.com")
		if err != nil {
			log.Errorf("Failed to create URL component %v", err)
			responses.PrivateInteractionResponse(s, i, "Something went wrong. Please try again later.")
			return
		}
		responses.PrivateInteractionResponseWithComponents(s, i, "Your account is not registered with Stablecog, click the link below to login/register.",
			[]discordgo.MessageComponent{
				urlComponent,
			})
		return
	}
}
