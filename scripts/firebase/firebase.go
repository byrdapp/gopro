package main

import (
	"flag"
	"fmt"

	"github.com/byrdapp/byrd-pro-api/internal/storage/firebase"
	"github.com/joho/godotenv"
)

func main() {
	env := flag.String("env", "development", "set environment")
	flag.Parse()
	initCreds(*env)

	fb, err := firebase.NewFB()
	if err != nil {
		panic(err)
	}
	storyId := []string{
		"-M05kMmz1y9Akej-_sQT",
		"-M05hhOJs3a4c-ecovbQ",
		"-LyixE0XBjlaPtvdIZZo",
		"-LoLVuxA43sEnzYOaIw8",
		"-LoiRQahmp98whji76vG",
		"-Lii2VmLIU2fD5zcDTRy",
		"-LihzSRNsavH2N_MJyuz",
		"-LhMgjwZleyJFMgf7dKq",
		"-LT2u16oPJnlKSfjvY2V",
		"-LT2sq1xWDK8z8FLUPEC",
		"-LSyta7jk2x7rZ6-LUIM",
		"-LTr-anTCguGtdZd9f7Q",
		"-LTr0N-mhcQipW3zqY44",
		"-Ly3k2MDShGG1l9ZJJAB",
		"-Ly3gatePv6j0Ovsru2K"}
	for _, id := range storyId {
		if err := fb.PutStoryData(id, "isFake", true); err != nil {
			panic(err)
		}
		fmt.Printf("Successfully updated FirebaseProfile with id %s\n", id)
	}
}

func initCreds(env string) {
	switch env {
	case "development":
		if err := godotenv.Load(".env"); err != nil {
			panic(err)
		}
		break
	case "production":
		if err := godotenv.Load("production.env"); err != nil {
			panic(err)
		}
		break
	default:
		panic("no evironment set")
	}
}
