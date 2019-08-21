package utils

import "strings"

// JoinStrings join strings
func JoinStrings(list []string) string {
	return strings.Join(list, ", ")
}
