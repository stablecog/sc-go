package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/deviceinfo"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

var MockRepo *Repository

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	ctx := context.Background()
	dbconn, err := database.GetSqlDbConn(!utils.GetEnv().GithubActions)
	if err != nil {
		log.Fatal("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	entClient, err := database.NewEntClient(dbconn)
	defer entClient.Close()
	if err != nil {
		log.Fatal("Failed to create ent client", "err", err)
		os.Exit(1)
	}

	// Redis setup
	origMockRedis := utils.GetEnv().MockRedis
	utils.GetEnv().MockRedis = true
	defer func() {
		utils.GetEnv().MockRedis = origMockRedis
	}()

	redis, err := database.NewRedis(ctx)
	if err != nil {
		log.Fatal("Error connecting to redis", "err", err)
		os.Exit(1)
	}

	//Create schema
	if err := entClient.Schema.Create(ctx); err != nil {
		log.Fatal("Failed to run migrations", "err", err)
		os.Exit(1)
	}

	MockRepo = &Repository{
		DB:       entClient,
		ConnInfo: dbconn,
		Redis:    redis,
		Ctx:      ctx,
	}

	// Create mockdata
	if err = MockRepo.CreateMockData(ctx); err != nil {
		log.Fatal("Failed to create mock data", "err", err)
		os.Exit(1)
	}

	return m.Run()
}

// Test that wrapper rolls back transaction when error is thrown
func TestTxWrapper(t *testing.T) {
	err := MockRepo.WithTx(func(tx *ent.Tx) error {
		DB := tx.Client()
		// Change something arbitrary
		_, err := DB.DeviceInfo.Create().SetType("rollback").SetOs("rollback").SetBrowser("rollback").Save(MockRepo.Ctx)
		assert.Nil(t, err)

		// Query to make sure exists
		dinfo := DB.DeviceInfo.Query().Where(deviceinfo.Type("rollback"), deviceinfo.Os("rollback"), deviceinfo.Browser("rollback")).FirstX(MockRepo.Ctx)
		assert.NotNil(t, dinfo)
		assert.Equal(t, "rollback", *dinfo.Type)

		// Throw an error to trigger rollback
		return errors.New("rollback")
	})

	assert.NotNil(t, err)
	// Should not be found
	_, err = MockRepo.DB.DeviceInfo.Query().Where(deviceinfo.Type("rollback"), deviceinfo.Os("rollback"), deviceinfo.Browser("rollback")).First(MockRepo.Ctx)
	assert.NotNil(t, err)
	assert.True(t, ent.IsNotFound(err))
}
