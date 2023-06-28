package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/discobot/components"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

var ErrNotAuthorized = errors.New("not authorized")

// Shared auth wrapper, returns nil if unauthorized
func (d *DiscoDomain) CheckAuthorization(s *discordgo.Session, i *discordgo.InteractionCreate) *ent.User {
	if i.Member == nil {
		return nil
	}
	u, err := d.Repo.GetUserByDiscordID(i.Member.User.ID)
	if err != nil && !ent.IsNotFound(err) {
		log.Errorf("Failed to get user by discord ID %v", err)
		responses.UnknownErrorPrivateInteractionResponse(s, i)
		return nil
	}
	if err != nil && ent.IsNotFound(err) {
		// Set token in redis
		token, err := d.Redis.SetDiscordVerifyToken(i.Member.User.ID)
		if err != nil {
			log.Errorf("Failed to set discord verify token in redis %v", err)
			responses.UnknownErrorPrivateInteractionResponse(s, i)
			return nil
		}

		urlComponent, err := components.AuthComponent("Sign in", fmt.Sprintf("https://stablecog.com/discord?token=%s&discord_id=%s", token, i.Member.User.ID))
		if err != nil {
			log.Errorf("Failed to create URL component %v", err)
			responses.UnknownErrorPrivateInteractionResponse(s, i)
			return nil
		}
		// Get duration as minutes
		responses.PrivateInteractionResponseWithComponents(s,
			i,
			"🔐 Authentication Required",
			"You must sign in to stablecog before you can use this command.\n\n",
			"By signing in you agree to our Terms of Service and Privacy Policy.",
			[]discordgo.MessageComponent{
				urlComponent,
			})
		// Delete message when link expires
		time.AfterFunc(shared.DISCORD_VERIFY_TOKEN_EXPIRY, func() {
			s.InteractionResponseDelete(i.Interaction)
		})
		return nil
	}

	return u
}
