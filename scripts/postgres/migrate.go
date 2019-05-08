package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/byblix/gopro/storage"
	firebase "github.com/byblix/gopro/storage/firebase"
	postgres "github.com/byblix/gopro/storage/postgres"
)

func main() {
	fmt.Println("Starting migration")

	if err := godotenv.Load(); err != nil {
		logrus.Errorf("Error cfg: %s", err)
	}

	// Init SQL db
	sqldb, err := postgres.NewPQ()
	if err != nil {
		logrus.Fatal(err)
	}

	// Get the FB profiles
	profiles, err := getProfilesFromFB()
	if err != nil {
		logrus.Errorf("Error getting profiles: %s", err)
	}

	err = insertProfilesSQL(sqldb, profiles)
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

func insertProfilesSQL(sqldb postgres.Service, profiles []*storage.Profile) error {
	var wg sync.WaitGroup
	if err := sqldb.Ping(); err != nil {
		logrus.Fatal(err)
	}
	ctx := context.Background()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, val := range profiles {
			if val.IsMedia {
				media := postgres.Media{
					Name:        fmt.Sprintf("%s %s", val.FirstName, val.LastName),
					DisplayName: val.DisplayName,
					UserID:      val.UserID,
					Email:       val.Email,
				}

				str, err := sqldb.CreateMedia(ctx, &media)
				if err != nil {
					logrus.Errorf("Didnt create row: %s", err)
				}
				fmt.Printf("Inserted this fellow: %s and created id in SQL: %v\n", media.DisplayName, str)
			} else {
				fmt.Printf("Skipping this guy %s\n", val.DisplayName)
			}
		}
	}()
	wg.Wait()

	return nil
}

// todo:
// func createMediatableDepartments(){}

// todo:
// func insertDepartmentSQL(){}
