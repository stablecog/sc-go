package responses

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/log"
)

type RESPONSE_PRIVACY int

const (
	PRIVATE RESPONSE_PRIVACY = iota
	PUBLIC
)

type InteractionResponseOptions struct {
	Content      *string
	EmbedTitle   string
	EmbedContent string
	EmbedFooter  string
	ImageURLs    []string
	Privacy      RESPONSE_PRIVACY
	Embeds       []*discordgo.MessageEmbed
	Components   []discordgo.MessageComponent
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
	for _, url := range options.ImageURLs {
		options.Embeds = append(options.Embeds, NewImageEmbed(url))
	}

	// Deref content
	var content string
	if options.Content != nil {
		content = *options.Content
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Flags:      flags,
			Embeds:     options.Embeds,
			Components: options.Components,
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

	// Create image embeds
	for _, url := range options.ImageURLs {
		options.Embeds = append(options.Embeds, NewImageEmbed(url))
	}

	// Embeds and components as pointers
	var embeds *[]*discordgo.MessageEmbed
	var components *[]discordgo.MessageComponent
	if len(options.Embeds) > 0 {
		embeds = &options.Embeds
	}
	if len(options.Components) > 0 {
		components = &options.Components
	}

	resp, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    options.Content,
		Components: components,
		Embeds:     embeds,
	})
	if err != nil {
		log.Errorf("Failed to edit interaction response: %v", err)
	}
	return resp, err
}

func ErrorResponseEdit(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.Message, error) {
	return InteractionEdit(s, i, &InteractionResponseOptions{
		EmbedTitle:   "😔",
		EmbedContent: "An unknown error occurred. Please try again later.",
		EmbedFooter:  "If this error persists, please contact the bot owner.",
	})
}

func InitialLoadingResponse(s *discordgo.Session, i *discordgo.InteractionCreate, privacy RESPONSE_PRIVACY) error {
	return InitialInteractionResponse(s, i, &InteractionResponseOptions{
		Privacy:    privacy,
		EmbedTitle: "⏱️ Working...Beep Boop",
	})
}
