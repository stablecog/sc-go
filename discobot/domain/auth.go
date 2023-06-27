package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/ent"
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

		urlComponent, err := responses.AuthComponent("Sign in", "Link Existing Account", fmt.Sprintf("https://stablecog.com/discordverify/%s", token))
		if err != nil {
			log.Errorf("Failed to create URL component %v", err)
			responses.UnknownErrorPrivateInteractionResponse(s, i)
			return nil
		}
		// Get duration as minutes
		responses.PrivateInteractionResponseWithComponents(s,
			i,
			"⚠️ Action Required",
			"You must authenticate your Discord account with Stablecog before you can use this command.\n\nYou can do this in by either by signing in with your Discord account or by connecting your discord account to an existing Stablecog account.",
			fmt.Sprintf("⏰ This link will expire in %d minutes.", int(shared.DISCORD_VERIFY_TOKEN_EXPIRY.Minutes())),
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
