package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBlacklisted(t *testing.T) {
	assert.True(t, IsBlacklisted("admin"))
	assert.True(t, IsBlacklisted("root"))
	assert.True(t, IsBlacklisted("administrator"))
	assert.True(t, IsBlacklisted("system"))
	assert.False(t, IsBlacklisted("joe_gitler"))
}
