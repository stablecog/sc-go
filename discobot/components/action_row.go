package components

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-multierror"
)

func NewActionRow(components ...*SCDiscordComponent) (discordgo.MessageComponent, error) {
	urlComponent := SCDiscordActionRow{
		Type:       discordgo.ActionsRowComponent,
		Components: components,
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
