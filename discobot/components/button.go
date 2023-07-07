package components

import "github.com/bwmarrin/discordgo"

// Creates a new action button
func NewButton(label string, id string, emoji string) *SCDiscordComponent {
	c := &SCDiscordComponent{
		Type:     discordgo.ButtonComponent,
		Style:    discordgo.SecondaryButton,
		Label:    label,
		CustomID: id,
	}
	if emoji != "" {
		c.Emoji = &discordgo.Emoji{
			Name: emoji,
		}
	}
	return c
}
