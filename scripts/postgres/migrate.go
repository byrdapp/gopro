package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/byblix/gopro/storage"
	firebase "github.com/byblix/gopro/storage/firebase"
	psqr "github.com/byblix/gopro/storage/postgres"
)

var psgr *psqr.Postgres

func main() {
	fmt.Println("Starting migration")

	if err := godotenv.Load(); err != nil {
		logrus.Errorf("Error cfg: %s", err)
	}
	// Get the FB profiles
	profiles, err := getProfilesFromFB()
	if err != nil {
		logrus.Errorf("Error getting profiles: %s", err)
	}

	err = insertProfilesSQL(profiles)
	if err != nil {
		logrus.Errorf("Error inserting profiles: %s", err)
	}

}

// ExportToPostgres -
func getProfilesFromFB() ([]*storage.Profile, error) {
	fbdb, err := firebase.InitFirebaseDB()
	if err != nil {
		return nil, err
	}
	prfs, err := fbdb.GetProfiles()
	if err != nil {
		return nil, err
	}
	return prfs, nil
}

func insertProfilesSQL(profiles []*storage.Profile) error {

	for _, val := range profiles[:3] {
		fmt.Println(val.DisplayName)
	}

	return nil

}
