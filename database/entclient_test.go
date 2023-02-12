package database

import (
	"testing"

	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewEntClient(t *testing.T) {
	dbconn, _ := GetSqlDbConn(utils.GetEnv("GITHUB_ACTIONS", "") != "true")

	client, err := NewEntClient(dbconn)
	assert.Nil(t, err)
	assert.NotNil(t, client)
}
