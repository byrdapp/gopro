package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/blixenkrone/gopro/utils/logger"

	"github.com/joho/godotenv"

	"github.com/blixenkrone/gopro/storage"
	firebase "github.com/blixenkrone/gopro/storage/firebase"
	postgres "github.com/blixenkrone/gopro/storage/postgres"
)

var log = logger.NewLogger()

func main() {
	fmt.Println("Starting migration")

	if err := godotenv.Load(); err != nil {
		log.Errorf("Error cfg: %s", err)
	}

	// Init SQL db
	sqldb, err := postgres.NewPQ()
	if err != nil {
		log.Fatal(err)
	}

	// Get the FB profiles
	profiles, err := getProfilesFromFB()
	if err != nil {
		log.Errorf("Error getting profiles: %s", err)
	}

	err = insertProfilesSQL(sqldb, profiles)
	if err != nil {
		log.Errorf("Error inserting profiles: %s", err)
	}

}

// ExportToPostgres -
func getProfilesFromFB() ([]*storage.FirebaseProfile, error) {
	fbdb, err := firebase.NewFB()
	if err != nil {
		return nil, err
	}
	prfs, err := fbdb.GetProfiles(context.Background())
	if err != nil {
		return nil, err
	}
	return prfs, nil
}

func insertProfilesSQL(sqldb storage.PQService, profiles []*storage.FirebaseProfile) error {
	var wg sync.WaitGroup
	if err := sqldb.Ping(); err != nil {
		log.Fatal(err)
	}
	// ctx := context.Background()
	for _, p := range profiles {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if p.IsProfessional {
				log.Println(p)

				// ! create profile with level and userUID
				// pro := storage.Professional{
				// 	ID: p.UserID,
				// }

				// _, err := sqldb.CreateProfessional(ctx, &pro)
				// if err != nil {
				// 	log.Errorf("Didnt create row: %s", err)
				// 	return
				// }
			}
		}()
		wg.Wait()
	}
	return nil
}

// TODO:

// todo:
// func importProfessionals(){}

// todo:
// func createMediatableDepartments(){}

// todo:
// func insertDepartmentSQL(){}
