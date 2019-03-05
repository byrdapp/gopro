package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	logger *log.Logger
)

// to run this locally with dev: $ go build && ./gopro -env="local"
// to run this locally with prod: $ go build && ./gopro -env="local-production"
func main() {

	logger := log.New(os.Stdout, "micro-out: ", log.LstdFlags|log.Lshortfile)
	if err := InitEnvironment(); err != nil {
		logger.Fatalln(err)
	}

	err := NewServer()
	if err != nil {
		logger.Fatalln(err)
	}
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
	return nil
}

// only needed when certain scripts must be run
func initScripts() {
	// Init DB:
	// db, err := storage.InitFirebaseDB()
	// scripts.WithdrawalsToCSV(db)
	// scripts.ProfilesToCSV(db)
}
