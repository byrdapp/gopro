package main

import (
	"flag"
	"os"

	"github.com/joho/godotenv"

	"github.com/byrdapp/byrd-pro-api/internal/server"
	"github.com/byrdapp/byrd-pro-api/pkg/logger"
)

var (
	local      = flag.Bool("local", false, "Do you want to run go run *.go with .env local file?")
	production = flag.Bool("production", false, "Is it production")
	mute       = flag.Bool("mute", false, "Mute public notificatons but not logging")
	log        = logger.NewLogger()
)

func init() {
	flag.Parse()
	if *local && !*production {
		if err := godotenv.Load(); err != nil {
			panic(err)
		}
		if *mute {
			os.Setenv("PANIC_NOTIFICATIONS", "false")
		}
		log.Infof("Running locally with %s env", os.Getenv("ENV"))
	}
}

func main() {

	srv, err := server.NewServer()
	if err != nil {
		panic(err)
	}

	srv.Routes()

	if err := srv.UseHTTP2(); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}
	srv.Infof("Serving on host w. address :3000")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	// * runs until os.SIGTERM happens
	srv.WaitForShutdown()
}
