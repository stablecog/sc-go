package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/stablecog/sc-go/shared"
)

// Remove + from email addresses
func RemovePlusFromEmail(email string) string {
	re := regexp.MustCompile(`\+[^)]*@`)
	return re.ReplaceAllString(email, "@")
}

// Extract integer from a string (ie. !tip 500 @bbedward -> 500)
var AmountAmbiguousError = fmt.Errorf("amount_ambiguous")
var AmountMissingError = fmt.Errorf("amount_not_found")
var AmountNotIntegerError = fmt.Errorf("amount_not_integer")

// Not actually using regex here, but it's a good place to put this function
func ExtractAmountsFromString(str string) (int, error) {
	newStr := RemoveLineBreaks(RemoveRedundantSpaces(str))
	splitStr := strings.Split(newStr, " ")
	matches := []string{}
	for _, s := range splitStr {
		// See if valid float
		_, err := strconv.ParseFloat(s, 64)
		if err == nil {
			matches = append(matches, s)
		}
	}

	if len(matches) > 1 {
		return 0, AmountAmbiguousError
	} else if len(matches) == 1 {
		// Convert to int
		amt, err := strconv.Atoi(strings.ReplaceAll(matches[0], " ", ""))
		if err != nil {
			return 0, AmountNotIntegerError
		}
		return amt, nil
	}
	return 0, AmountMissingError
}

// Validate username
var (
	UsernameLengthError           = errors.New("username_length_error")
	UsernameStartsWithLetterError = errors.New("username_must_start_with_a_letter")
	UsernameCharError             = errors.New("username_can_only_contain_letters_numbers_hyphens")
	UsernameProfaneError          = errors.New("username_profanity")
	UsernameBlacklistedError      = errors.New("username_blacklisted")
)

func IsValidUsername(username string) error {
	// Rule 1: Must be between 3 and 25 characters
	if len(username) < 3 || len(username) > 25 {
		return UsernameLengthError
	}

	// Rule 2: Must start with a letter
	if !isLetter(username[0]) {
		return UsernameStartsWithLetterError
	}

	// Rule 3: Must contain only letters or numbers
	matched, err := regexp.MatchString("^[a-zA-Z0-9-]+$", username)
	if err != nil {
		// If there's an error in regex matching, return a generic error
		return errors.New("username validation failed")
	}
	if !matched {
		return UsernameCharError
	}

	// Rule 4: Can't be profane
	if goaway.IsProfane(username) {
		return UsernameProfaneError
	}

	// Rule 5: Can't be blacklisted
	if shared.IsBlacklisted(username) {
		return UsernameBlacklistedError
	}

	return nil // Username is valid, return nil error
}

func isLetter(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z')
}
