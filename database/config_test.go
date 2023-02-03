package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSqlDbConnPostgres(t *testing.T) {
	// Postgres
	os.Setenv("POSTGRES_DB", "pippin")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PASSWORD", "password")
	os.Setenv("POSTGRES_HOST", "127.0.0.1")
	defer os.Unsetenv("POSTGRES_DB")
	defer os.Unsetenv("POSTGRES_USER")
	defer os.Unsetenv("POSTGRES_PASSWORD")
	defer os.Unsetenv("POSTGRES_HOST")

	conn, err := GetSqlDbConn()
	assert.Nil(t, err)

	assert.Equal(t, "postgres://user:password@127.0.0.1:5432/pippin", conn.DSN())
	assert.Equal(t, "pgx", conn.Dialect())
}
