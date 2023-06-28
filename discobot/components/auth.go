package components

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-multierror"
)

const AuthComponentSignInID = "sc_discord_sign_in"

type SCDiscordActionRow struct {
	Type discordgo.ComponentType `json:"type"`
	// Components is a slice of MessageComponents
	Components []SCDiscordComponent `json:"components"`
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

func AuthComponent(label string, url string) (discordgo.MessageComponent, error) {
	urlComponent := SCDiscordActionRow{
		Type: discordgo.ActionsRowComponent,
		Components: []SCDiscordComponent{
			{
				Type:  discordgo.ButtonComponent,
				Style: discordgo.LinkButton,
				Label: label,
				URL:   url,
			},
			{
				Type:  discordgo.ButtonComponent,
				Style: discordgo.LinkButton,
				Label: "Terms of Service",
				URL:   "https://stablecog.com/terms",
			},
			{
				Type:  discordgo.ButtonComponent,
				Style: discordgo.LinkButton,
				Label: "Privacy Policy",
				URL:   "https://stablecog.com/privacy",
			},
		},
	}
	// Marshal
	var mErr *multierror.Error
	b, err := json.Marshal(urlComponent)
	mErr = multierror.Append(mErr, err)
	messageComponent, err := discordgo.MessageComponentFromJSON(b)
	mErr = multierror.Append(mErr, err)
	if mErr.ErrorOrNil() != nil {
		return nil, mErr
	}
	return messageComponent, nil
}
