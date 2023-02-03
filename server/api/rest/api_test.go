// Test setup for the controller package
package rest

import (
	"context"
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
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
	ctx := context.Background()
	// Setup embedded postgres
	postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Username("test").
		Password("test").
		Database("test").
		Version(embeddedpostgres.V14))
	err := postgres.Start()
	if err != nil {
		klog.Fatalf("Failed to start embedded postgres: %v", err)
		os.Exit(1)
	}
	defer postgres.Stop()

	// Set in env
	os.Setenv("POSTGRES_DB", "test")
	os.Setenv("POSTGRES_USER", "test")
	os.Setenv("POSTGRES_PASSWORD", "test")
	os.Setenv("POSTGRES_HOST", "localhost")
	defer os.Unsetenv("POSTGRES_DB")
	defer os.Unsetenv("POSTGRES_USER")
	defer os.Unsetenv("POSTGRES_PASSWORD")
	defer os.Unsetenv("POSTGRES_HOST")
	dbconn, err := database.GetSqlDbConn()
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
