package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/blixenkrone/byrd/byrd-pro-api/pkg/logger"

	"database/sql"

	"github.com/joho/godotenv"

	"github.com/blixenkrone/byrd/byrd-pro-api/internal/storage"
	firebase "github.com/blixenkrone/byrd/byrd-pro-api/internal/storage/firebase"
	postgres "github.com/blixenkrone/byrd/byrd-pro-api/internal/storage/postgres"
)

var log = logger.NewLogger()

func main() {
	fmt.Println("Starting migration")

	if err := godotenv.Load(); err != nil {
		log.Errorf("Error cfg: %s", err)
	}

	// Init SQL db
	db, err := sql.Open("postgres", "connstr")
	if err != nil {
		panic(err)
	}
	sqldb := postgres.New(db)

	_ = sqldb

	// Get the FB profiles

	// err = insertProfilesSQL(sqldb, profiles)
	// if err != nil {
	// 	log.Errorf("Error inserting profiles: %s", err)
	// }

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

				// _, err := sqldb(ctx, &pro)
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
