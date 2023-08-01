package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemovePlusFromEmail(t *testing.T) {
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello@stablecog.com"))
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello+123@stablecog.com"))
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello+1@stablecog.com"))
	assert.Equal(t, "hello@stablecog.com", RemovePlusFromEmail("hello+abcdef@stablecog.com"))
	assert.Equal(t, "hello", RemovePlusFromEmail("hello"))
}

func TestExtractAmountsFromString(t *testing.T) {
	amt, err := ExtractAmountsFromString("hello 123 45")
	assert.Equal(t, AmountAmbiguousError, err)
	amt, err = ExtractAmountsFromString("hello 123.45")
	assert.Equal(t, AmountNotIntegerError, err)
	amt, err = ExtractAmountsFromString("hello world")
	assert.Equal(t, AmountMissingError, err)
	amt, err = ExtractAmountsFromString("!hello 123 here is for gas")
	assert.NoError(t, err)
	assert.Equal(t, 123, amt)
}

func TestIsValidUsername(t *testing.T) {
	usernames := []string{
		"john-doe",
		"johndoe123",
		"j-doe",
		"j_doe",
		"johndoe12345678901234567",
		"johndoe",
		"-john_doe",
		"john-do_e",
		"a",
		"123456789123456789123456788999",
		"fuckyou",
		"This-sh1t",
		"admin",
	}

	assert.Nil(t, IsValidUsername(usernames[0]))
	assert.Nil(t, IsValidUsername(usernames[1]))
	assert.Nil(t, IsValidUsername(usernames[2]))
	assert.ErrorIs(t, UsernameCharError, IsValidUsername(usernames[3]))
	assert.Nil(t, IsValidUsername(usernames[4]))
	assert.Nil(t, IsValidUsername(usernames[5]))
	assert.ErrorIs(t, UsernameStartsWithLetterError, IsValidUsername(usernames[6]))
	assert.ErrorIs(t, UsernameCharError, IsValidUsername(usernames[7]))
	assert.ErrorIs(t, UsernameLengthError, IsValidUsername(usernames[8]))
	assert.ErrorIs(t, UsernameLengthError, IsValidUsername(usernames[9]))
	assert.ErrorIs(t, UsernameProfaneError, IsValidUsername(usernames[10]))
	assert.ErrorIs(t, UsernameProfaneError, IsValidUsername(usernames[11]))
	assert.ErrorIs(t, UsernameBlacklistedError, IsValidUsername(usernames[12]))
}
