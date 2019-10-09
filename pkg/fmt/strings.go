package utils

import (
	"strconv"
	"strings"
)

// JoinStrings join strings
func JoinStrings(list []string) string {
	return strings.Join(list, ", ")
}

// ParseBool parses string to bool val
func ParseBool(stringValToBool string) (bool, error) {
	return strconv.ParseBool(stringValToBool)
}
