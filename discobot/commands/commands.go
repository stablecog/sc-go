package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/discobot/domain"
)

// Create new wrapper and register interactions/commands
func NewDiscordCommandWrapper(repo *repository.Repository, redis *database.RedisWrapper) *DiscordCommandWrapper {
	// Create wrapper
	wrapper := &DiscordCommandWrapper{
		Disco: &domain.DiscoDomain{Repo: repo, Redis: redis},
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
	Disco    *domain.DiscoDomain
	Commands []*DiscordCommand
}

// Specification for specific commands
type DiscordCommand struct {
	ApplicationCommand *discordgo.ApplicationCommand
	Handler            func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
