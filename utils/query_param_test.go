package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddQueryParam(t *testing.T) {
	url := "https://stablecog.com"

	urlWithParam, err := AddQueryParam(url, QueryParam{Key: "hello", Value: "stablecog"})
	assert.Nil(t, err)
	assert.Equal(t, "https://stablecog.com?hello=stablecog", urlWithParam)

	urlWithParam, err = AddQueryParam(urlWithParam, QueryParam{Key: "hello2", Value: "stablecog2"})
	assert.Nil(t, err)
	assert.Equal(t, "https://stablecog.com?hello=stablecog&hello2=stablecog2", urlWithParam)
}
