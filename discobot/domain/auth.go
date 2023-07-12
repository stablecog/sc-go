package domain

import (
	"errors"
	"fmt"
	"net/url"
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
	var discordId string
	var discordUsername string
	var avatarUrl string
	if i.Member != nil {
		discordId = i.Member.User.ID
		discordUsername = i.Member.User.Username
		avatarUrl = i.Member.AvatarURL("128")
	} else {
		discordId = i.User.ID
		discordUsername = i.User.Username
		avatarUrl = i.User.AvatarURL("128")
	}
	u, err := d.Repo.GetUserByDiscordID(discordId)
	if err != nil && !ent.IsNotFound(err) {
		log.Errorf("Failed to get user by discord ID %v", err)
		responses.ErrorResponseInitial(s, i, responses.PRIVATE)
		return nil
	}
	if err != nil && ent.IsNotFound(err) {
		// Set token in redis
		token, err := d.Redis.SetDiscordVerifyToken(discordId)
		if err != nil {
			log.Errorf("Failed to set discord verify token in redis %v", err)
			responses.ErrorResponseInitial(s, i, responses.PRIVATE)
			return nil
		}

		// Create URL params for login
		params := url.Values{}
		params.Add("platform_token", token)
		params.Add("platform_user_id", discordId)
		params.Add("platform_username", discordUsername)
		params.Add("platform_avatar_url", avatarUrl)

		// Auth msg
		err = responses.InitialInteractionResponse(s,
			i,
			&responses.InteractionResponseOptions{
				EmbedTitle:   "🚀 Sign in to start",
				EmbedContent: "Create a Stablecog account or sign in to your existing one to start using the bot.\n\n",
				EmbedFooter:  "By signing in you agree to our Terms of Service and Privacy Policy.",
				ActionRowOne: []*components.SCDiscordComponent{
					components.NewLinkButton("Sign in", fmt.Sprintf("https://stablecog.com/connect/discord?%s", params.Encode()), "🔑"),
					components.NewLinkButton("Terms & Policies", "https://stablecog.com/legal", "📜"),
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
