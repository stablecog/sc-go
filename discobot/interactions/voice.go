package interactions

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/discobot/components"
	"github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	srvresponses "github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

func (c *DiscordInteractionWrapper) NewVoiceoverCommand() *DiscordInteraction {
	// Get voiceover speakers from DB
	// Discord has a limit of 25 choices per command
	speakers, err := c.Repo.GetVoiceverSpeakersWithName(25)
	if err != nil {
		log.Errorf("Error getting voiceover speakers: %v", err)
		panic(err)
	}
	// Sort speakers by locale
	sort.Slice(speakers, func(i, j int) bool {
		return speakers[i].Locale < speakers[j].Locale
	})
	// Sort so english is at the top
	sort.Slice(speakers, func(i, j int) bool {
		return speakers[i].Locale == "en"
	})
	// Move isDefault to the top
	for i, speaker := range speakers {
		if speaker.IsDefault {
			(speakers) = append([]*ent.VoiceoverSpeaker{speaker}, append((speakers)[:i], (speakers)[i+1:]...)...)
			break
		}
	}
	// Create speakers as discord choices
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, speaker := range speakers {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  fmt.Sprintf("%s (%s)", *speaker.Name, speaker.Locale),
			Value: speaker.ID.String(),
		})
	}
	return &DiscordInteraction{
		// Command spec
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "speak",
			Description: "Create a voiceover with Stablecog.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: "The prompt for the voiceover.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "speaker",
					Description: "The speaker for the voiceover.",
					Required:    false,
					Choices:     choices,
				},
			},
		},
		// The handler for the command
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var discordUserId string
			if i.Member != nil {
				discordUserId = i.Member.User.ID
			} else {
				discordUserId = i.User.ID
			}
			if u := c.Disco.CheckAuthorization(s, i); u != nil {
				// Access options in the order provided by the user.
				options := i.ApplicationCommandData().Options

				// Parse options
				var prompt string
				var speaker *uuid.UUID

				for _, option := range options {
					switch option.Name {
					case "prompt":
						prompt = option.StringValue()
					case "speaker":
						speaker = utils.ToPtr(uuid.MustParse(option.StringValue()))
					}
				}

				req := requests.CreateVoiceoverRequest{
					Prompt:    prompt,
					SpeakerId: speaker,
				}
				credits, err := c.Repo.GetNonExpiredCreditTotalForUser(u.ID, nil)
				if err != nil {
					log.Errorf("Error getting credits for user: %v", err)
					responses.ErrorResponseInitial(s, i, responses.PRIVATE)
					return
				}
				if credits < int(req.Cost()) {
					responses.InitialInteractionResponse(s, i, responses.InsufficientCreditsResponseOptions(req.Cost(), int32(credits)))
					return
				}

				// Always create initial message
				responses.InitialLoadingResponse(s, i, responses.PUBLIC)

				// Create voiceover
				res, _, wErr := c.SCWorker.CreateVoiceover(
					enttypes.SourceTypeDiscord,
					nil,
					u,
					nil,
					req,
				)
				if wErr != nil || len(res.Outputs) == 0 {
					if errors.Is(wErr.Err, srvresponses.InsufficientCreditsErr) {
						credits, err := c.Repo.GetNonExpiredCreditTotalForUser(u.ID, nil)
						if err != nil {
							log.Errorf("Error getting credits for user: %v", err)
							responses.ErrorResponseEdit(s, i)
							return
						}
						responses.InteractionEdit(s, i, responses.InsufficientCreditsResponseOptions(req.Cost(), int32(credits)))
						return
					}
					log.Errorf("Error creating voiceover: %v", wErr.Err)
					responses.ErrorResponseEdit(s, i)
					return
				}

				var videoUrls []string
				for _, output := range res.Outputs {
					if output.VideoFileURL != nil {
						videoUrls = append(videoUrls, *output.VideoFileURL)
					}
				}
				if len(videoUrls) == 0 {
					log.Errorf("Error creating voiceover: %v", err)
					responses.ErrorResponseEdit(s, i)
					return
				}

				// Send the image
				_, err = responses.InteractionEdit(s, i, &responses.InteractionResponseOptions{
					Content: utils.ToPtr(fmt.Sprintf("<@%s> **%s**\n%s", discordUserId, prompt, videoUrls[0])),
					Embeds:  nil,
					ActionRowOne: []*components.SCDiscordComponent{
						components.NewLinkButton("MP3", *res.Outputs[0].AudioFileURL, "🎵"),
						components.NewLinkButton("MP4", *res.Outputs[0].VideoFileURL, "🎥"),
					},
				},
				)
				if err != nil {
					log.Error(err)
					responses.ErrorResponseEdit(s, i)
				}

				err = c.Repo.UpdateLastSeenAt(u.ID)
				if err != nil {
					log.Warn("Error updating last seen at", "err", err, "user", u.ID.String())
				}
			} else {
				c.LoginInteractionMap.Put(discordUserId, &LoginInteraction{
					Session:     s,
					Interaction: i,
					InsertedAt:  time.Now(),
				})
			}
		},
	}
}
