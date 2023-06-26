package responses

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-multierror"
)

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

func URLComponent(label string, url string) (discordgo.MessageComponent, error) {
	urlComponent := SCDiscordActionRow{
		Type: discordgo.ActionsRowComponent,
		Components: []SCDiscordComponent{
			{
				Type:  discordgo.ButtonComponent,
				Style: discordgo.LinkButton,
				Label: label,
				URL:   url,
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
