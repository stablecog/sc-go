package responses

import "github.com/bwmarrin/discordgo"

const EMBED_PURPLE = 11437547

func NewEmbed(title string, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color:       EMBED_PURPLE,
		Title:       title,
		Description: description,
	}
}
