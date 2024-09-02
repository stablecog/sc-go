package main

// asynq consumer
import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/quecon/processor"
	"github.com/stablecog/sc-go/shared"
)

var Version = "dev"
var CommitMsg = "dev"

func usage() {
	fmt.Printf("Usage %s [options]\n", os.Args[0])
	flag.PrintDefaults()
	return
}

func main() {
	log.Infof("SC QueCon %s", Version)

	// Close loki if exists
	defer log.CloseLoki()

	showHelp := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *showHelp {
		usage()
		os.Exit(0)
	}

	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Warn("Error loading .env file (this is fine)", "err", err)
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

	// Setup asynq server
	options := redis.Client.Options()
	redisOptions := asynq.RedisClientOpt{
		Addr: options.Addr,
		DB:   options.DB,
	}

	// Setup handler wrapper
	processor := processor.NewQueueProcessor()

	srv := asynq.NewServer(
		redisOptions,
		asynq.Config{
			Concurrency: 5,
			Queues:      shared.ASYNQ_QUEUE_DEFINITIONS,
		},
	)

	// Define handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(shared.ASYNQ_TASK_GENERATE, processor.HandleImageJob)

	if err := srv.Run(mux); err != nil {
		log.Fatal("Error running asynq server", "err", err)
	}
}
