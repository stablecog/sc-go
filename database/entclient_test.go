package database

import (
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/stretchr/testify/assert"
	"k8s.io/klog/v2"
)

func TestNewEntClient(t *testing.T) {
	// Setup embedded postgres
	postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Username("test2").
		Password("test2").
		Database("test2").
		Port(12345).
		Version(embeddedpostgres.V14))
	err := postgres.Start()
	if err != nil {
		klog.Fatalf("Failed to start embedded postgres: %v", err)
		os.Exit(1)
	}
	defer postgres.Stop()

	// Set in env
	os.Setenv("POSTGRES_DB", "test2")
	os.Setenv("POSTGRES_USER", "test2")
	os.Setenv("POSTGRES_PASSWORD", "test2")
	os.Setenv("POSTGRES_PORT", "12345")
	os.Setenv("POSTGRES_HOST", "localhost")
	defer os.Unsetenv("POSTGRES_DB")
	defer os.Unsetenv("POSTGRES_USER")
	defer os.Unsetenv("POSTGRES_PASSWORD")
	defer os.Unsetenv("POSTGRES_PORT")
	defer os.Unsetenv("POSTGRES_HOST")
	dbconn, _ := GetSqlDbConn()

	client, err := NewEntClient(dbconn)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}
