package storage

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

// DBInstance -
type DBInstance struct {
	Client  *db.Client
	Context context.Context
}

// InitFirebaseDB SE
func InitFirebaseDB() (*DBInstance, error) {
	ctx := context.Background()
	config := &firebase.Config{
		DatabaseURL: os.Getenv("FB_DATABASE_URL"),
	}
	jsonPath := "fb-" + os.Getenv("ENV") + ".json"
	opt := option.WithCredentialsJSON(GetAWSSecrets(jsonPath))
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		panic(err)
	}
	client, err := app.Database(ctx)
	if err != nil {
		return nil, err
	}

	return &DBInstance{
		Client:  client,
		Context: ctx,
	}, nil
}

// GetTransactions - this guy
func GetTransactions(db *DBInstance) ([]*Transaction, error) {
	p := os.Getenv("ENV") + "/transactions"
	transaction := []*Transaction{}
	fmt.Printf("Path: %s\n", p)
	ref := db.Client.NewRef(p)
	if err := ref.Get(db.Context, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

// GetWithdrawals - this guy
func GetWithdrawals(db *DBInstance) ([]*Withdrawals, error) {
	p := os.Getenv("ENV") + "/transactions"
	wd := []*Withdrawals{}
	fmt.Printf("Path: %s\n", p)
	ref := db.Client.NewRef(p)
	if err := ref.Get(db.Context, wd); err != nil {
		return nil, err
	}
	return wd, nil
}

// GetProfile get a single profile instance
func GetProfile(db *DBInstance, uid string) (*Profile, error) {
	path := os.Getenv("ENV") + "/profiles" + uid
	prf := Profile{}
	fmt.Printf("Path: %s\n", path)
	ref := db.Client.NewRef(path)
	if err := ref.Get(db.Context, &path); err != nil {
		return nil, err
	}
	return &prf, nil
}

// GetProfiles get a single profile instance
func GetProfiles(db *DBInstance) ([]*Profile, error) {
	path := os.Getenv("ENV") + "/profiles"
	prfs := []*Profile{}
	fmt.Printf("Path: %s\n", path)
	ref := db.Client.NewRef(path)
	if err := ref.Get(db.Context, &path); err != nil {
		return nil, err
	}
	return prfs, nil
}
