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
		responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		return nil
	}
	if err != nil && ent.IsNotFound(err) {
		// Set token in redis
		token, err := d.Redis.SetDiscordVerifyToken(i.Member.User.ID)
		if err != nil {
			log.Errorf("Failed to set discord verify token in redis %v", err)
			responses.ErrorResponseInitial(s, i, responses.PRIVATE)
			return nil
		}

		// Get duration as minutes
		err = responses.InitialInteractionResponse(s,
			i,
			&responses.InteractionResponseOptions{
				EmbedTitle:   "🔐 Authentication Required",
				EmbedContent: "You must sign in to stablecog before you can use this command.\n\n",
				EmbedFooter:  "By signing in you agree to our Terms of Service and Privacy Policy.",
				ActionRowOne: []*components.SCDiscordComponent{
					components.NewLinkButton("Sign in", fmt.Sprintf("https://stablecog.com/discordlink?token=%s&discord_id=%s", token, i.Member.User.ID)),
					components.NewLinkButton("Terms of Service", "https://stablecog.com/terms"),
					components.NewLinkButton("Privacy Policy", "https://stablecog.com/privacy"),
				},
				Privacy: responses.PRIVATE,
			},
		)
		if err != nil {
			responses.ErrorResponseInitial(s, i, responses.PRIVATE)
			return nil
		}
		// Delete message when link expires
		time.AfterFunc(shared.DISCORD_VERIFY_TOKEN_EXPIRY, func() {
			s.InteractionResponseDelete(i.Interaction)
		})
		return nil
	}

	return u
}
