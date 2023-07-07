package components

import "github.com/bwmarrin/discordgo"

type SCDiscordActionRow struct {
	Type discordgo.ComponentType `json:"type"`
	// Components is a slice of MessageComponents
	Components []*SCDiscordComponent `json:"components"`
}

type SCDiscordComponent struct {
	Type     discordgo.ComponentType `json:"type"`
	Style    discordgo.ButtonStyle   `json:"style"`
	Label    string                  `json:"label,omitempty"`
	Emoji    *discordgo.Emoji        `json:"emoji,omitempty"`
	CustomID string                  `json:"custom_id,omitempty"`
	URL      string                  `json:"url,omitempty"`
	Disabled *bool                   `json:"disabled,omitempty"`
}
