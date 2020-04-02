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
		"-M3Gh0heU45IXJ22lDtb",
		"-M3GjKpTl1Vg0KheDq1Q",
		"-M3GnHUt8hpgAGvwHfxV",
		"-M3H3uMTswT2NJuko5gS",
		"-M3HJ9CroiJ_ISLagr3_",
		"-M3HLQdpjNYI-bC3e1S7",
		"-M3HOKJwhwNilZvYtybT",
		"-M3HS9EaXgfdeygf6Kfn",
		"-M3HS9EaXgfdeygf6Kfn",
		"-M3HUMPG84hawlUHG-vV",
	}
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
	case "production":
		if err := godotenv.Load("production.env"); err != nil {
			panic(err)
		}
	default:
		panic("no evironment set")
	}
}
