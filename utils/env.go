package utils

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
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

// InitEnvironment : set the cli flag -env=local if must be run locally
func InitEnvironment() error {
	env := LookupEnv("ENV", "development")
	flag.StringVar(&env, "env", env, "Environment used")
	flag.Parse()

	if env == "local" {
		if err := godotenv.Load(); err != nil {
			return err
		}
	} else {
		fmt.Println("Server CFG is being used")
	}

	fmt.Printf("%s environment is used as config\n", env)
	return nil
}
