package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/blixenkrone/gopro/storage"
	aws "github.com/blixenkrone/gopro/storage/aws"

	"google.golang.org/api/iterator"

	"firebase.google.com/go/auth"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

// Firebase is the created struct for creating FB method refs
type Firebase struct {
	Client  *db.Client
	Auth    *auth.Client
	Context context.Context
}

// Service contains the methods attached to the Firebase struct
type Service interface {
	GetTransactions() ([]*storage.Transaction, error)
	GetWithdrawals() ([]*storage.Withdrawals, error)
	GetProfile(uid string) (*storage.Profile, error)
	GetProfiles() ([]*storage.Profile, error)
	UpdateData(uid string, prop string, value string) error
	GetAuth() ([]*auth.ExportedUserRecord, error)
	DeleteAuthUserByUID(uid string) error
	GetToken(ctx context.Context, uid string) (string, error)
	VerifyToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// New SE
func New() (Service, error) {
	ctx := context.Background()
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

	return &Firebase{
		Client:  client,
		Context: ctx,
		Auth:    auth,
	}, nil
}

// VerifyToken verify JWT handled by middleware.go returning the uid
func (db *Firebase) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := db.Auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	sub := token.Subject
	uid := token.UID
	spew.Dump(sub)
	spew.Dump(uid)
	return token, nil
}

// GetTransactions - this guy
func (db *Firebase) GetTransactions() ([]*storage.Transaction, error) {
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
func (db *Firebase) GetWithdrawals() ([]*storage.Withdrawals, error) {
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
func (db *Firebase) GetProfile(uid string) (*storage.Profile, error) {
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
func (db *Firebase) GetProfiles() ([]*storage.Profile, error) {
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
func (db *Firebase) UpdateData(uid string, prop string, value string) error {
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
func (db *Firebase) GetAuth() ([]*auth.ExportedUserRecord, error) {
	// path := os.Getenv("ENV")
	profiles := []*auth.ExportedUserRecord{}
	iter := db.Auth.Users(db.Context, "")
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
func (db *Firebase) DeleteAuthUserByUID(uid string) error {
	err := db.Auth.DeleteUser(db.Context, uid)
	if err != nil {
		return err
	}
	return nil
}

// GetToken returns token as a string
func (db *Firebase) GetToken(ctx context.Context, uid string) (string, error) {
	return db.Auth.CustomToken(ctx, uid)
}
