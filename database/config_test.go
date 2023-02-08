package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSqlDbConnPostgres(t *testing.T) {
	// Get original env vars
	postgresDb := os.Getenv("POSTGRES_DB")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresHost := os.Getenv("POSTGRES_HOST")
	// Postgres
	os.Setenv("POSTGRES_DB", "pippin")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PASSWORD", "password")
	os.Setenv("POSTGRES_HOST", "127.0.0.1")
	// Reset env
	defer func() {
		os.Setenv("POSTGRES_DB", postgresDb)
		os.Setenv("POSTGRES_USER", postgresUser)
		os.Setenv("POSTGRES_PASSWORD", postgresPassword)
		os.Setenv("POSTGRES_HOST", postgresHost)
	}()

	conn, err := GetSqlDbConn(false)
	assert.Nil(t, err)

	assert.Equal(t, "postgres://user:password@127.0.0.1:5432/pippin", conn.DSN())
	assert.Equal(t, "pgx", conn.Dialect())
}

func TestGetSqlDbConnMock(t *testing.T) {
	conn, err := GetSqlDbConn(true)
	assert.Nil(t, err)

	assert.Equal(t, "file:testing?cache=shared&mode=memory&_fk=1", conn.DSN())
	assert.Equal(t, "sqlite3", conn.Dialect())
}
