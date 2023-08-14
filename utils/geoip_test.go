package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCountryFromIP(t *testing.T) {
	db, err := NewGeoIPService(true)
	assert.Nil(t, err)
	defer db.Close()

	country, err := db.GetCountryFromIP("81.2.69.142")
	assert.Nil(t, err)
	assert.Equal(t, "GB", country)
}
