package repository

import (
	"context"
	"os"
	"testing"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

var MockRepo *Repository

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	ctx := context.Background()
	dbconn, err := database.GetSqlDbConn(utils.GetEnv("GITHUB_ACTIONS", "") != "true")
	if err != nil {
		klog.Fatalf("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	entClient, err := database.NewEntClient(dbconn)
	defer entClient.Close()
	if err != nil {
		klog.Fatalf("Failed to create ent client: %v", err)
		os.Exit(1)
	}

	// Redis setup
	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")

	redis, err := database.NewRedis(ctx)
	if err != nil {
		klog.Fatalf("Error connecting to redis: %v", err)
		os.Exit(1)
	}

	//Create schema
	if err := entClient.Schema.Create(ctx); err != nil {
		klog.Fatalf("Failed to run migrations: %v", err)
		os.Exit(1)
	}

	MockRepo = &Repository{
		DB:    entClient,
		Redis: redis,
		Ctx:   ctx,
	}

	// Create mockdata
	if err = MockRepo.CreateMockData(ctx); err != nil {
		klog.Fatalf("Failed to create mock data: %v", err)
		os.Exit(1)
	}

	return m.Run()
}
