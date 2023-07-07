package components

import "github.com/bwmarrin/discordgo"

// Creates a new link button
func NewLinkButton(label string, url string) *SCDiscordComponent {
	return &SCDiscordComponent{
		Type:  discordgo.ButtonComponent,
		Style: discordgo.LinkButton,
		Label: label,
		URL:   url,
	}
}
