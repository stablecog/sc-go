package responses

import "github.com/bwmarrin/discordgo"

const EMBED_PURPLE = 11437547

func NewEmbed(title, description, footer string) *discordgo.MessageEmbed {
	var footerEmbed *discordgo.MessageEmbedFooter
	if footer != "" {
		footerEmbed = &discordgo.MessageEmbedFooter{
			Text: footer,
		}
	}
	return &discordgo.MessageEmbed{
		Color:       EMBED_PURPLE,
		Title:       title,
		Description: description,
		Footer:      footerEmbed,
	}
}
