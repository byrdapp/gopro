package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func LookupEnv(val, fallback string) string {
	v, ok := os.LookupEnv(val)
	if !ok {
		if fallback != "" {
			log.Printf("Error getting env, using: %s", fallback)
			return fallback
		}
		log.Error("Error getting fallback")
	}
	return v
}
