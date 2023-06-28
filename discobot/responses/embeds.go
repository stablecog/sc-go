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

func NewImageEmbed(url string) *discordgo.MessageEmbed {
	image := &discordgo.MessageEmbedImage{
		URL: url,
	}
	return &discordgo.MessageEmbed{
		Color: EMBED_PURPLE,
		URL:   "https://stablecog.com",
		Image: image,
	}
}
