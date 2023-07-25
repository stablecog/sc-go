package utils

import (
	"regexp"
)

// Remove + from email addresses
func RemovePlusFromEmail(email string) string {
	re := regexp.MustCompile(`\+[^)]*@`)
	return re.ReplaceAllString(email, "@")
}
