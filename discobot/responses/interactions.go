package responses

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/discobot/components"
	"github.com/stablecog/sc-go/log"
)

type RESPONSE_PRIVACY int

const (
	PRIVATE RESPONSE_PRIVACY = iota
	PUBLIC
)

type InteractionResponseOptions struct {
	Content        *string
	EmbedTitle     string
	EmbedContent   string
	EmbedFooter    string
	ImageURLs      []string
	ImageURLsEmbed []string
	VideoURLs      []string
	Privacy        RESPONSE_PRIVACY
	Embeds         []*discordgo.MessageEmbed
	ActionRowOne   []*components.SCDiscordComponent
}

// For the first response to an interaction
func InitialInteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, options *InteractionResponseOptions) error {
	// Catch panics in discordgo response
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Panic caught in interaction response: %v", err)
		}
	}()

	var flags discordgo.MessageFlags
	if options.Privacy == PRIVATE {
		flags = discordgo.MessageFlagsEphemeral
	}

	// Create generic default embed
	if options.EmbedTitle != "" || options.EmbedContent != "" || options.EmbedFooter != "" {
		options.Embeds = append(options.Embeds, NewEmbed(options.EmbedTitle, options.EmbedContent, options.EmbedFooter))
	}

	// Create image embeds
	for _, url := range options.ImageURLsEmbed {
		options.Embeds = append(options.Embeds, NewImageEmbed(url))
	}

	// Deref content
	var content string
	if options.Content != nil {
		content = *options.Content
	}

	// ! TODO ?
	// Would be nice to be able to attach files without downloading and uploading them to discord
	files := make([]*discordgo.File, len(options.ImageURLs)+len(options.VideoURLs))
	for i, url := range options.ImageURLs {
		response, err := http.Get(url)
		if err != nil {
			log.Errorf("Failed to get image: %v", err)
			return err
		}
		defer response.Body.Close()

		files[i] = &discordgo.File{
			Name:        url,
			ContentType: "image/jpeg",
			Reader:      response.Body,
		}
	}

	for i, url := range options.VideoURLs {
		response, err := http.Get(url)
		if err != nil {
			log.Errorf("Failed to get video: %v", err)
			return err
		}
		defer response.Body.Close()
		files[i+len(options.ImageURLs)] = &discordgo.File{
			Name:        url,
			ContentType: "video/mp4",
			Reader:      response.Body,
		}
	}

	// Create components
	discComponents := []discordgo.MessageComponent{}
	if len(options.ActionRowOne) > 0 {
		actionRowOne, err := components.NewActionRow(options.ActionRowOne...)
		if err != nil {
			log.Errorf("Failed to create action row: %v", err)
			return err
		}
		discComponents = append(discComponents, actionRowOne)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Flags:      flags,
			Embeds:     options.Embeds,
			Components: discComponents,
			Files:      files,
		},
	})
	if err != nil {
		log.Errorf("Failed to respond to interaction: %v", err)
	}
	return err
}

// For edits to an interaction response
func InteractionEdit(s *discordgo.Session, i *discordgo.InteractionCreate, options *InteractionResponseOptions) (*discordgo.Message, error) {
	// Catch panics in discordgo response
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Panic caught in interaction edit: %v", err)
		}
	}()

	// Create generic default embed
	if options.EmbedTitle != "" || options.EmbedContent != "" || options.EmbedFooter != "" {
		options.Embeds = append(options.Embeds, NewEmbed(options.EmbedTitle, options.EmbedContent, options.EmbedFooter))
	}

	// Embeds and components as pointers
	embeds := []*discordgo.MessageEmbed{}
	if len(options.Embeds) > 0 {
		embeds = options.Embeds
	}

	// Create image embeds
	for _, url := range options.ImageURLsEmbed {
		embeds = append(options.Embeds, NewImageEmbed(url))
	}

	// ! TODO ?
	// Would be nice to be able to attach files without downloading and uploading them to discord
	files := make([]*discordgo.File, len(options.ImageURLs)+len(options.VideoURLs))
	for i, url := range options.ImageURLs {
		response, err := http.Get(url)
		if err != nil {
			log.Errorf("Failed to get image: %v", err)
			return nil, err
		}
		defer response.Body.Close()

		files[i] = &discordgo.File{
			Name:        url,
			ContentType: "image/jpeg",
			Reader:      response.Body,
		}
	}

	for i, url := range options.VideoURLs {
		response, err := http.Get(url)
		if err != nil {
			log.Errorf("Failed to get video: %v", err)
			return nil, err
		}
		defer response.Body.Close()
		files[i+len(options.ImageURLs)] = &discordgo.File{
			Name:        url,
			ContentType: "video/mp4",
			Reader:      response.Body,
		}
	}

	// Create components
	discComponents := []discordgo.MessageComponent{}
	if len(options.ActionRowOne) > 0 {
		actionRowOne, err := components.NewActionRow(options.ActionRowOne...)
		if err != nil {
			log.Errorf("Failed to create action row: %v", err)
			return nil, err
		}
		discComponents = append(discComponents, actionRowOne)
	}

	resp, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    options.Content,
		Components: &discComponents,
		Embeds:     &embeds,
		Files:      files,
	})
	if err != nil {
		log.Errorf("Failed to edit interaction response: %v", err)
	}
	return resp, err
}

func ErrorResponseInitial(s *discordgo.Session, i *discordgo.InteractionCreate, privacy RESPONSE_PRIVACY) error {
	return InitialInteractionResponse(s, i, &InteractionResponseOptions{
		EmbedTitle:   "😔",
		EmbedContent: "An unknown error occurred. Please try again later.",
		EmbedFooter:  "If this error persists, please contact the bot owner.",
		Privacy:      privacy,
	})
}

func ErrorResponseEdit(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.Message, error) {
	return InteractionEdit(s, i, &InteractionResponseOptions{
		EmbedTitle:   "😔",
		EmbedContent: "An unknown error occurred. Please try again later.",
		EmbedFooter:  "If this error persists, please contact the bot owner.",
	})
}

func ErrorResponseEditValidation(s *discordgo.Session, i *discordgo.InteractionCreate, content string) (*discordgo.Message, error) {
	return InteractionEdit(s, i, &InteractionResponseOptions{
		EmbedTitle:   "🚫",
		EmbedContent: content,
	})
}

func ErrorResponseInitialValidation(s *discordgo.Session, i *discordgo.InteractionCreate, content string, privacy RESPONSE_PRIVACY) error {
	return InitialInteractionResponse(s, i, &InteractionResponseOptions{
		EmbedTitle:   "🚫",
		EmbedContent: content,
		Privacy:      privacy,
	})
}

func InitialLoadingResponse(s *discordgo.Session, i *discordgo.InteractionCreate, privacy RESPONSE_PRIVACY) error {
	return InitialInteractionResponse(s, i, &InteractionResponseOptions{
		Privacy:    privacy,
		EmbedTitle: "<a:loading:1128598014597017680>  Working on it",
	})
}

// When a user does not have enough credits to perform an action
func InsufficientCreditsResponseOptions(needed, have int32) *InteractionResponseOptions {
	return &InteractionResponseOptions{
		EmbedTitle:   "Insufficient credits",
		EmbedContent: fmt.Sprintf("You need %d credits to perform this action, but you only have %d. ", needed, have),
		EmbedFooter:  "You can subscribe or purchase additional credits at any time.",
		ActionRowOne: []*components.SCDiscordComponent{
			components.NewLinkButton("Pricing Page", "https://stablecog.com/pricing", "🪙"),
		},
		Privacy: PRIVATE,
	}
}
