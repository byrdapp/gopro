package utils

import (
	"os"

	"github.com/blixenkrone/gopro/pkg/logger"
)

var log = logger.NewLogger()

// LookupEnv -
func LookupEnv(val, fallback string) string {
	v, ok := os.LookupEnv(val)
	if !ok {
		if fallback != "" {
			log.Printf("Error getting env, using: %s", fallback)
			return fallback
		}
	}
	return v
}
