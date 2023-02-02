// Test setup for the controller package
package rest

import (
	"context"
	"os"
	"testing"

	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/api/websocket"
	"github.com/stablecog/go-apps/shared"
	"k8s.io/klog/v2"
)

// A valid websocket ID that will be acceptable by APIs
const MockWSId = "e08abf9698f7d27e634de0d36cc974a0d908ec41c0a7e5e5738d2431f9a700e3"

var MockController *RestAPI

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	// We use an in-memory sqlite database for testing
	ctx := context.Background()
	dbconn, err := database.GetSqlDbConn(true)
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

	//Create schema
	if err := entClient.Schema.Create(ctx); err != nil {
		klog.Fatalf("Failed to run migrations: %v", err)
		os.Exit(1)
	}
	repo := &repository.Repository{
		DB:  entClient,
		Ctx: ctx,
	}

	// Mock data
	if err := database.CreateMockData(ctx, entClient, repo); err != nil {
		klog.Fatalf("Failed to create mock data: %v", err)
		os.Exit(1)
	}

	// Populate cache
	if err := repo.UpdateCache(); err != nil {
		klog.Fatalf("Failed to populate cache: %v", err)
		os.Exit(1)
	}

	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")

	redis, err := database.NewRedis(ctx)
	if err != nil {
		klog.Fatalf("Error connecting to redis: %v", err)
		os.Exit(1)
	}

	// Setup fake websocket hub
	hub := websocket.NewHub()
	go hub.Run()
	// Add user to hub
	hub.Register <- &websocket.Client{
		Uid: MockWSId,
	}

	// Setup controller
	MockController = &RestAPI{
		Repo:                       repo,
		Redis:                      redis,
		Hub:                        hub,
		CogRequestWebsocketConnMap: shared.NewSyncMap[string](),
	}

	return m.Run()
}
