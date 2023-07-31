package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	goaway "github.com/TwiN/go-away"
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
	UsernameLengthError           = errors.New("username length must be between 3 and 25 characters")
	UsernameStartsWithLetterError = errors.New("username must start with a letter")
	UsernameCharError             = errors.New("username can only contain letters or numbers")
	UsernameHyphenError           = errors.New("username can't contain both hyphens and underscores")
	UsernameProfaneError          = errors.New("username contains profane words")
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
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	if err != nil {
		// If there's an error in regex matching, return a generic error
		return errors.New("username validation failed")
	}
	if !matched {
		return UsernameCharError
	}

	// Rule 4: Can contain hyphens or underscores, but not both
	hasHyphen := false
	hasUnderscore := false
	for _, char := range username {
		if char == '-' {
			hasHyphen = true
		} else if char == '_' {
			hasUnderscore = true
		}
	}

	if hasHyphen && hasUnderscore {
		return UsernameHyphenError
	}

	if goaway.IsProfane(username) {
		return UsernameProfaneError
	}

	return nil // Username is valid, return nil error
}

func isLetter(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z')
}
