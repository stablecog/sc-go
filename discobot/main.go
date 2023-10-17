package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/discobot/interactions"
	dresponses "github.com/stablecog/sc-go/discobot/responses"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/analytics"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/shared/queue"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
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

	// Setup qdrant
	qdrantClient, err := qdrant.NewQdrantClient(ctx)
	if err != nil {
		log.Fatal("Error connecting to qdrant", "err", err)
		os.Exit(1)
	}

	// Create repository (database access)
	repo := &repository.Repository{
		DB:       entClient,
		ConnInfo: dbconn,
		Qdrant:   qdrantClient,
		Redis:    redis,
		Ctx:      ctx,
	}

	// Sync map for tracking requests
	sMap := shared.NewSyncMap[chan requests.CogWebhookMessage]()

	// Make a sync map for tracking login requests
	loginInteractionMap := shared.NewSyncMap[*interactions.LoginInteraction]()

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
	// Also delete records older than 10 minutes from loginInteractionMap
	cronSscheduler.Every(10).Minutes().Do(func() {
		items := loginInteractionMap.GetAll()
		for k, v := range items {
			if time.Since(v.InsertedAt) > 10*time.Minute {
				loginInteractionMap.Delete(k)
			}
		}
	})
	// Safety checker
	safetyChecker := utils.NewTranslatorSafetyChecker(ctx, os.Getenv("OPENAI_API_KEY"), false)

	// Create analytics service
	analyticsService := analytics.NewAnalyticsService()
	defer analyticsService.Close()

	// Setup rabbitmq client
	rabbitmqClient, err := queue.NewRabbitMQClient(ctx, os.Getenv("RABBITMQ_AMQP_URL"))
	if err != nil {
		log.Fatalf("Error connecting to rabbitmq: %v", err)
	}
	defer rabbitmqClient.Close()

	// Setup interactions
	cmdWrapper := interactions.NewDiscordInteractionWrapper(repo, redis, database.NewSupabaseAuth(), sMap, qThrottler, safetyChecker, analyticsService, loginInteractionMap, rabbitmqClient)

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Register messageCreate as a callback for the messageCreate events.
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		// check if the message is "!tip"
		if strings.HasPrefix(m.Content, "!tip") {
			cmdWrapper.HandleTip(s, m)
		}
	})

	// Intents
	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Remove stale commands
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Fatalf("Cannot get existing commands: %v", err)
	}
	wantedCommandNames := make([]string, len(cmdWrapper.Commands))
	for i, v := range cmdWrapper.Commands {
		wantedCommandNames[i] = v.ApplicationCommand.Name
	}
	for _, v := range existingCommands {
		if slices.Contains(wantedCommandNames, v.Name) {
			continue
		}
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
		}
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
			if strings.HasPrefix(i.MessageComponentData().CustomID, "upscale:") {
				// String should look like this: upscale:%s:number:%d
				splitStr := strings.Split(i.MessageComponentData().CustomID, ":")
				if len(splitStr) != 4 {
					log.Error("Invalid upscale custom id, parsing length", "custom_id", i.MessageComponentData().CustomID)
					return
				}
				// Parse uuid
				outputId, err := uuid.Parse(splitStr[1])
				if err != nil {
					log.Error("Invalid upscale custom id, parsing uuid", "custom_id", i.MessageComponentData().CustomID)
					return
				}
				// Get number
				number, err := strconv.Atoi(splitStr[3])
				if err != nil {
					log.Error("Invalid upscale custom id, parsing number", "custom_id", i.MessageComponentData().CustomID)
					return
				}
				cmdWrapper.HandleUpscaleGeneration(s, i, outputId, number)
			} else {
				handler := cmdWrapper.GetHandlerForComponent(i.MessageComponentData().CustomID)
				if handler != nil {
					handler(s, i)
				}
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

	// This channel tells the bot when a user has authenticated with Stablecog
	// so we can send them a message
	stableCogAuthChannel := redis.Client.Subscribe(ctx, shared.REDIS_DISCORD_COG_CHANNEL)
	defer stableCogAuthChannel.Close()

	// Start SSE redis subscription
	go func() {
		log.Info("Listening for stablecog auth messages", "channel", shared.REDIS_DISCORD_COG_CHANNEL)
		for msg := range stableCogAuthChannel.Channel() {
			var authMsg responses.DiscordRedisStreamMessage
			err := json.Unmarshal([]byte(msg.Payload), &authMsg)
			if err != nil {
				log.Error("Error unmarshalling sc auth message", "err", err)
				continue
			}

			dmUser, err := s.User(authMsg.DiscordId)
			if err != nil {
				log.Error("Error getting user information", "err", err)
				continue
			}

			dmChannel, err := s.UserChannelCreate(authMsg.DiscordId)
			if err != nil {
				log.Error("Error creating dm channel", "err", err)
				continue
			}
			s.ChannelMessageSendEmbed(dmChannel.ID, dresponses.NewEmbed(fmt.Sprintf("👋 Hi, @%s!", dmUser.Username), "I'm Stuart, the Stablecog bot. I'm here to provide you a suite of AI tools.\n\nTry one of the following commands:\n\n/imagine\n/speak", ""))

			// See if interaction exists in sync map too
			i := loginInteractionMap.Get(authMsg.DiscordId)
			if i != nil {
				// Update interaction
				dresponses.InteractionEdit(i.Session, i.Interaction, &dresponses.InteractionResponseOptions{
					EmbedTitle:   "✅ Authenticated!",
					EmbedContent: "You have successfully authenticated with Stablecog! You can now use various commands.\n\nTry one of the following commands:\n\n/imagine\n/speak",
				})

				// Remove from sync map
				loginInteractionMap.Delete(authMsg.DiscordId)
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	log.Info("Press Ctrl+C to exit")
	<-stop

	log.Info("Gracefully shutting down.")
}
