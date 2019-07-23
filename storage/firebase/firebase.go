package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/blixenkrone/gopro/storage"
	aws "github.com/blixenkrone/gopro/storage/aws"

	"google.golang.org/api/iterator"

	"firebase.google.com/go/auth"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

// DBInstance is the created struct for creating FB method refs
type DBInstance struct {
	Client  *db.Client
	Context context.Context
	Auth    *auth.Client
}

var ctx = context.Background()

// InitFirebaseDB SE
func InitFirebaseDB() (*DBInstance, error) {
	config := &firebase.Config{
		DatabaseURL: os.Getenv("FB_DATABASE_URL"),
	}
	jsonPath := "fb-" + os.Getenv("ENV") + ".json"
	opt := option.WithCredentialsJSON(aws.GetAWSSecrets(jsonPath))
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		return nil, err
	}
	client, err := app.Database(ctx)
	if err != nil {
		return nil, err
	}
	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return &DBInstance{
		Client:  client,
		Context: ctx,
		Auth:    auth,
	}, nil
}

// GetTransactions - this guy
func (db *DBInstance) GetTransactions() ([]*storage.Transaction, error) {
	p := os.Getenv("ENV") + "/transactions"
	transaction := []*storage.Transaction{}
	fmt.Printf("Path: %s\n", p)
	ref := db.Client.NewRef(p)
	if err := ref.Get(db.Context, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

// GetWithdrawals - this guy
func (db *DBInstance) GetWithdrawals() ([]*storage.Withdrawals, error) {
	p := os.Getenv("ENV") + "/transactions"
	wd := []*storage.Withdrawals{}
	fmt.Printf("Path: %s\n", p)
	ref := db.Client.NewRef(p)
	if err := ref.Get(db.Context, wd); err != nil {
		return nil, err
	}
	return wd, nil
}

// GetProfile get a single profile instance
func (db *DBInstance) GetProfile(uid string) (*storage.Profile, error) {
	path := os.Getenv("ENV") + "/profiles/" + uid
	prf := storage.Profile{}
	fmt.Printf("Path: %s\n", path)
	ref := db.Client.NewRef(path)
	if err := ref.Get(db.Context, &prf); err != nil {
		return nil, err
	}
	return &prf, nil
}

// GetProfiles get multiple profile instances
func (db *DBInstance) GetProfiles() ([]*storage.Profile, error) {
	var prfs []*storage.Profile
	path := os.Getenv("ENV") + "/profiles"
	ref := db.Client.NewRef(path)
	res, err := ref.OrderByKey().GetOrdered(db.Context)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Path: %s\n", ref.Path)
	for _, r := range res {
		var p *storage.Profile
		if err := r.Unmarshal(&p); err != nil {
			return nil, err
		}
		prfs = append(prfs, p)
	}
	return prfs, nil
}

// UpdateData userID is the uid to change profile to. Prop and value is a map.
func (db *DBInstance) UpdateData(uid string, prop string, value string) error {
	data := make(map[string]interface{})
	data[prop] = value
	path := os.Getenv("ENV") + "/profiles/" + uid
	ref := db.Client.NewRef(path)
	fmt.Println("Path to set:", ref.Path)
	err := ref.Update(db.Context, data)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully updated profile with %s\n", data[prop])
	return nil
}

// GetAuth -
func (db *DBInstance) GetAuth() ([]*auth.ExportedUserRecord, error) {
	// path := os.Getenv("ENV")
	profiles := []*auth.ExportedUserRecord{}
	iter := db.Auth.Users(ctx, "")
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, user)
	}
	return profiles, nil
}

// DeleteAuthUserByUID -
func (db *DBInstance) DeleteAuthUserByUID(uid string) error {
	err := db.Auth.DeleteUser(ctx, uid)
	if err != nil {
		return err
	}
	return nil
}
