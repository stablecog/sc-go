package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/discobot/interactions"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

var Version = "dev"

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer
)

func main() {
	log.Infof("Starting SC Discobot v%v", Version)
	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("Error loading .env file (this is fine)", "err", err)
	}

	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	ctx := context.Background()

	// Setup redis
	redis, err := database.NewRedis(ctx)
	if err != nil {
		log.Fatal("Error connecting to redis", "err", err)
		os.Exit(1)
	}

	// Setup sql
	log.Info("🏡 Connecting to database...")
	dbconn, err := database.GetSqlDbConn(false)
	if err != nil {
		log.Fatal("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	entClient, err := database.NewEntClient(dbconn)
	if err != nil {
		log.Fatal("Failed to create ent client", "err", err)
		os.Exit(1)
	}
	defer entClient.Close()
	// Run migrations
	// We can't run on supabase, :(
	if utils.GetEnv("RUN_MIGRATIONS", "") == "true" {
		log.Info("🦋 Running migrations...")
		if err := entClient.Schema.Create(ctx); err != nil {
			log.Fatal("Failed to run migrations", "err", err)
			os.Exit(1)
		}
	}

	// Create repository (database access)
	repo := &repository.Repository{
		DB:       entClient,
		ConnInfo: dbconn,
		Redis:    redis,
		Ctx:      ctx,
	}

	// Sync map for tracking requests
	sMap := shared.NewSyncMap[chan requests.CogWebhookMessage]()

	// Q Throttler
	qThrottler := shared.NewQueueThrottler(ctx, redis.Client, shared.REQUEST_COG_TIMEOUT)

	// Get models, schedulers and put in cache
	log.Info("📦 Populating cache...")
	err = repo.UpdateCache()
	if err != nil {
		// ! Not getting these is fatal and will result in crash
		panic(err)
	}
	// Update periodically
	cronSscheduler := gocron.NewScheduler(time.UTC)
	cronSscheduler.Every(5).Minutes().StartAt(time.Now().Add(5 * time.Minute)).Do(func() {
		log.Info("📦 Updating cache...")
		err = repo.UpdateCache()
		if err != nil {
			log.Error("Error updating cache", "err", err)
		}
	})

	// Setup interactions
	cmdWrapper := interactions.NewDiscordInteractionWrapper(repo, redis, database.NewSupabaseAuth(), sMap, qThrottler)

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Info("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(cmdWrapper.Commands))
	for i, v := range cmdWrapper.Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v.ApplicationCommand)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", v.ApplicationCommand.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// Register handlers
	// Components are part of interactions, so we register InteractionCreate handler
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			handler := cmdWrapper.GetHandlerForCommand(i.ApplicationCommandData().Name)
			if handler != nil {
				handler(s, i)
			}
		case discordgo.InteractionMessageComponent:
			handler := cmdWrapper.GetHandlerForComponent(i.MessageComponentData().CustomID)
			if handler != nil {
				handler(s, i)
			}
		}
	})

	defer s.Close()

	// This redis subscription has the following purpose:
	// For API token requests, they are synchronous with API requests
	// so we need to send the response back to the appropriate channel
	apiTokenChannel := redis.Client.Subscribe(ctx, shared.REDIS_APITOKEN_COG_CHANNEL)
	defer apiTokenChannel.Close()

	// Start SSE redis subscription
	go func() {
		log.Info("Listening for api messages", "channel", shared.REDIS_APITOKEN_COG_CHANNEL)
		for msg := range apiTokenChannel.Channel() {
			var cogMessage requests.CogWebhookMessage
			err := json.Unmarshal([]byte(msg.Payload), &cogMessage)
			if err != nil {
				log.Error("Error unmarshalling cog webhook message", "err", err)
				continue
			}

			if chl := sMap.Get(cogMessage.Input.ID.String()); chl != nil {
				chl <- cogMessage
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Info("Press Ctrl+C to exit")
	<-stop

	log.Info("Removing commands...")
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Info("Gracefully shutting down.")
}
