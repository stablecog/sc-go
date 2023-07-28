package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
