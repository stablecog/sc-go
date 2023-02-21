// Sets up a CLI to trigger the various cron jobs
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/stablecog/sc-go/cron/jobs"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"k8s.io/klog/v2"
)

func usage() {
	fmt.Printf("Usage %s [options]\n", os.Args[0])
	flag.PrintDefaults()
	return
}

func main() {
	showHelp := flag.Bool("help", false, "Show help")
	healthCheck := flag.Bool("healthCheck", false, "Run the health check job")
	syncMeili := flag.Bool("syncMeili", false, "Sync the meili index")
	stats := flag.Bool("stats", false, "Run the stats job")
	allJobs := flag.Bool("all", false, "Run all jobs in a blocking process")
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Set("v", "3")

	flag.Parse()

	if *showHelp {
		usage()
		os.Exit(0)
	}

	// Load .env
	err := godotenv.Load("../.env")
	if err != nil {
		klog.Warningf("Error loading .env file (this is fine): %v", err)
	}

	ctx := context.Background()

	// Setup redis
	redis, err := database.NewRedis(ctx)
	if err != nil {
		klog.Fatalf("Error connecting to redis: %v", err)
		os.Exit(1)
	}

	// Setup sql
	klog.Infoln("🏡 Connecting to database...")
	dbconn, err := database.GetSqlDbConn(false)
	if err != nil {
		klog.Fatalf("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	entClient, err := database.NewEntClient(dbconn)
	if err != nil {
		klog.Fatalf("Failed to create ent client: %v", err)
		os.Exit(1)
	}
	defer entClient.Close()

	// Create repostiory
	// Create repository (database access)
	repo := &repository.Repository{
		DB:    entClient,
		Redis: redis,
		Ctx:   ctx,
	}

	// Create a job runner
	jobRunner := jobs.JobRunner{
		Repo:  repo,
		Redis: redis,
		Ctx:   ctx,
		Meili: database.NewMeiliSearchClient(),
	}

	if *healthCheck {
		err := jobRunner.CheckHealth()
		if err != nil {
			klog.Fatalf("Error running health check: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *syncMeili {
		err := jobRunner.SyncMeili()
		if err != nil {
			klog.Fatalf("Error syncing meili: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *stats {
		err := jobRunner.GetAndSetStats()
		if err != nil {
			klog.Fatalf("Error running stats job: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *allJobs {
		klog.Infoln("🏡 Starting all jobs...")
		s := gocron.NewScheduler(time.UTC)
		s.Every(60).Seconds().Do(jobRunner.SyncMeili)
		s.Every(60).Seconds().Do(jobRunner.GetAndSetStats)
		s.StartBlocking()
		os.Exit(0)
	}

	// Generic path, they didn't say what they wanted.
	usage()
	os.Exit(1)

}
