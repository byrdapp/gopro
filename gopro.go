package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// to run this locally with dev: $ go build && ./gopro -env="local"
// to run this locally with prod: $ go build && ./gopro -env="local-production"
func main() {

	if err := InitEnvironment(); err != nil {
		log.Fatalln(err)
	}

	if err := NewServer(); err != nil {
		log.Fatalln(err)
	}

	// initScripts()
}

// InitEnvironment : set the cli flag -env=local if must be run locally
func InitEnvironment() error {
	env := os.Getenv("ENV")
	flag.StringVar(&env, "env", env, "Environment used")
	flag.Parse()
	if env == "local" {
		if err := godotenv.Load(); err != nil {
			return err
		}
	}
	fmt.Println(os.Getenv("ENV") + " " + "is running")
	return nil
}

// only needed when certain scripts must be run
func initScripts() {

	// if err := scripts.DeleteUnusedAuthProfiles(); err != nil {
	// 	log.Fatalf("Error in mail: %s", err)
	// }
	// if err := scripts.ChangeProfileUserPicture(); err != nil {
	// 	log.Fatalf("Error in main: %s", err)
	// }
	// scripts.WithdrawalsToCSV(db)
	// scripts.ProfilesToCSV(db)
}
