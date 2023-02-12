// Test setup for the controller package
package rest

import (
	"context"
	"os"
	"testing"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/server/api/sse"
	"github.com/stablecog/sc-go/utils"
	"k8s.io/klog/v2"
)

// A valid sse stream ID that will be acceptable by APIs
const MockSSEId = "e08abf9698f7d27e634de0d36cc974a0d908ec41c0a7e5e5738d2431f9a700e3"

var MockController *RestAPI

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

	repo := &repository.Repository{
		DB:    entClient,
		Redis: redis,
		Ctx:   ctx,
	}

	// Mock data
	if err := repo.CreateMockData(ctx); err != nil {
		klog.Fatalf("Failed to create mock data: %v", err)
		os.Exit(1)
	}

	// Populate cache
	if err := repo.UpdateCache(); err != nil {
		klog.Fatalf("Failed to populate cache: %v", err)
		os.Exit(1)
	}

	// Setup fake sse hub
	hub := sse.NewHub(redis, repo)
	go hub.Run()
	// Add user to hub
	hub.Register <- &sse.Client{
		Uid:  MockSSEId,
		Send: make(chan []byte, 256),
	}

	// Setup controller
	MockController = &RestAPI{
		Repo:  repo,
		Redis: redis,
		Hub:   hub,
	}

	return m.Run()
}
