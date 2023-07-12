package interactions

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/discobot/responses"
)

func (c *DiscordInteractionWrapper) NewHelpCommand() *DiscordInteraction {
	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "help",
			Description: "Show help information about using this bot.",
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			responses.InitialInteractionResponse(s, i, &responses.InteractionResponseOptions{
				Privacy: responses.PRIVATE,
				Embeds: []*discordgo.MessageEmbed{
					responses.NewHelpEmbed(),
				},
			})
		},
	}
}
