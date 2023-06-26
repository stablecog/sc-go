package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database/repository"
)

// Create new wrapper and register interactions/commands
func NewDiscordCommandWrapper(repo *repository.Repository) *DiscordCommandWrapper {
	// Create wrapper
	wrapper := &DiscordCommandWrapper{
		Repo: repo,
	}
	// Register commands
	commands := []*DiscordCommand{
		wrapper.NewAuthenticateCommand(),
	}
	// Set commands
	wrapper.Commands = commands
	return wrapper
}

// Wrapper for all commands
type DiscordCommandWrapper struct {
	Repo     *repository.Repository
	Commands []*DiscordCommand
}

// Specification for specific commands
type DiscordCommand struct {
	ApplicationCommand *discordgo.ApplicationCommand
	Handler            func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
