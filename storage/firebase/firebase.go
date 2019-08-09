package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/blixenkrone/gopro/utils/logger"

	"github.com/blixenkrone/gopro/storage"
	aws "github.com/blixenkrone/gopro/storage/aws"

	"google.golang.org/api/iterator"

	"firebase.google.com/go/auth"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

var log = logger.NewLogger()

// Firebase is the created struct for creating FB method refs
type Firebase struct {
	Client  *db.Client
	Auth    *auth.Client
	Context context.Context // context.Backgroun() - use r.Context()
}

// Service contains the methods attached to the Firebase struct
type Service interface {
	GetTransactions() ([]*storage.Transaction, error)
	GetWithdrawals() ([]*storage.Withdrawals, error)
	GetProfile(ctx context.Context, uid string) (*storage.FirebaseProfile, error)
	GetProfileByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	GetProfiles() ([]*storage.FirebaseProfile, error)
	UpdateData(uid string, prop string, value string) error
	GetAuth() ([]*auth.ExportedUserRecord, error)
	DeleteAuthUserByUID(uid string) error
	CreateCustomToken(ctx context.Context, uid string) (string, error)
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

// GetProfile get a single FirebaseProfile instance
func (db *Firebase) GetProfile(ctx context.Context, uid string) (*storage.FirebaseProfile, error) {
	path := os.Getenv("ENV") + "/profiles/" + uid
	prf := storage.FirebaseProfile{}
	fmt.Printf("Path: %s\n", path)
	ref := db.Client.NewRef(path)
	_, err := ref.GetWithETag(ctx, &prf)
	if err != nil {
		return nil, err
	}
	return &prf, nil
}

// GetProfileByEmail returns single UserRecord instance from email
func (db *Firebase) GetProfileByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	usr, err := db.Auth.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return usr, nil
}

// GetProfiles get multiple FirebaseProfile instances
func (db *Firebase) GetProfiles() ([]*storage.FirebaseProfile, error) {
	var prfs []*storage.FirebaseProfile
	path := os.Getenv("ENV") + "/profiles"
	ref := db.Client.NewRef(path)
	res, err := ref.OrderByKey().GetOrdered(db.Context)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Path: %s\n", ref.Path)
	for _, r := range res {
		var p *storage.FirebaseProfile
		if err := r.Unmarshal(&p); err != nil {
			return nil, err
		}
		prfs = append(prfs, p)
	}
	return prfs, nil
}

// UpdateData userID is the uid to change FirebaseProfile to. Prop and value is a map.
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
	fmt.Printf("Successfully updated FirebaseProfile with %s\n", data[prop])
	return nil
}

// GetAuth -
func (db *Firebase) GetAuth() ([]*auth.ExportedUserRecord, error) {
	// path := os.Getenv("ENV")
	profile := []*auth.ExportedUserRecord{}
	iter := db.Auth.Users(db.Context, "")
	for {
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		profile = append(profile, user)
	}
	return profile, nil
}

// DeleteAuthUserByUID -
func (db *Firebase) DeleteAuthUserByUID(uid string) error {
	err := db.Auth.DeleteUser(db.Context, uid)
	if err != nil {
		return err
	}
	return nil
}

// CreateCustomToken returns token as a string
func (db *Firebase) CreateCustomToken(ctx context.Context, uid string) (string, error) {
	return db.Auth.CustomToken(ctx, uid)
}

// VerifyToken verify JWT handled by middleware.go returning the uid
func (db *Firebase) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	t, err := db.Auth.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// IsAdmin returns token as a string
func (db *Firebase) IsAdmin(ctx context.Context, uid string) (bool, error) {
	user, err := db.Auth.GetUser(ctx, uid)
	if err != nil {
		log.Fatal(err)
	}
	// The claims can be accessed on the user record.
	admin, ok := user.CustomClaims["admin"]
	if !ok {
		var err error
		err = fmt.Errorf("Error getting admin Claims")
		return false, err
	}
	return admin.(bool), nil
}
