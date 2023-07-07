package components

import "github.com/bwmarrin/discordgo"

// Creates a new link button
func NewLinkButton(label string, url string, emoji string) *SCDiscordComponent {
	c := &SCDiscordComponent{
		Type:  discordgo.ButtonComponent,
		Style: discordgo.LinkButton,
		Label: label,
		URL:   url,
	}
	if emoji != "" {
		c.Emoji = &discordgo.Emoji{
			Name: emoji,
		}
	}
	return c
}
