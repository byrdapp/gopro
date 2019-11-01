package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/internal/storage"
	aws "github.com/blixenkrone/gopro/internal/storage/aws"

	"github.com/blixenkrone/gopro/pkg/logger"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/db"
)

var log = logger.NewLogger()

// Firebase is the created struct for creating FB method refs
type Firebase struct {
	Client  *db.Client
	Auth    *auth.Client
	Context context.Context // context.Backgroun() - use r.Context()
}

// ! Get profile params to switch profile type (reg, media, pro)
// ! Integrate GET's from FB to .go

// NewFB SE
func NewFB() (storage.FBService, error) {
	ctx := context.Background()
	config := &firebase.Config{
		DatabaseURL: os.Getenv("FB_DATABASE_URL"),
	}
	jsonPath := "fb-" + os.Getenv("ENV") + ".json"
	opt := option.WithCredentialsJSON(aws.GetAWSSecrets(jsonPath))
	// opt := option.WithCredentialsJSON(p)
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

	log.Infoln("Started Firebase admin")

	return &Firebase{
		Client:  client,
		Context: ctx,
		Auth:    auth,
	}, nil
}

// UpdateData userID is the uid to change FirebaseProfile to. Prop and value is a map.
func (db *Firebase) UpdateData(uid string, prop string, value string) error {
	data := make(map[string]interface{})
	data[prop] = value
	path := os.Getenv("ENV") + "/profiles/" + uid
	ref := db.Client.NewRef(path)
	err := ref.Update(db.Context, data)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully updated FirebaseProfile with %s\n", data[prop])
	return nil
}

// GetTransactions - this guy
func (db *Firebase) GetTransactions() ([]*storage.Transaction, error) {
	p := os.Getenv("ENV") + "/transactions"
	transaction := []*storage.Transaction{}
	ref := db.Client.NewRef(p)
	if err := ref.Get(db.Context, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

// GetWithdrawals - this guy
func (db *Firebase) GetWithdrawals(ctx context.Context) ([]*storage.Withdrawals, error) {
	p := os.Getenv("ENV") + "/withdrawals"
	wd := []*storage.Withdrawals{}
	ref := db.Client.NewRef(p)
	res, err := ref.OrderByKey().GetOrdered(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting ordered error")
	}
	for _, r := range res {
		w := &storage.Withdrawals{}
		if err := r.Unmarshal(w); err != nil {
			return nil, errors.Wrap(err, "unmarshall struct error")
		}
		wd = append(wd, w)
	}
	return wd, nil
}

// GetProfile get a single FirebaseProfile instance
func (db *Firebase) GetProfile(ctx context.Context, uid string) (*storage.FirebaseProfile, error) {
	path := os.Getenv("ENV") + "/profiles"
	prf := storage.FirebaseProfile{}
	ref := db.Client.NewRef(path).Child(uid)
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
func (db *Firebase) GetProfiles(ctx context.Context) ([]*storage.FirebaseProfile, error) {
	var prfs []*storage.FirebaseProfile
	path := os.Getenv("ENV") + "/profiles"
	ref := db.Client.NewRef(path)
	res, err := ref.OrderByKey().GetOrdered(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "reference error")
	}
	fmt.Printf("Path: %s\n", ref.Path)
	for _, r := range res {
		var p storage.FirebaseProfile
		if err := r.Unmarshal(&p); err != nil {
			return nil, errors.Wrap(err, "unmarshall struct error")
		}
		prfs = append(prfs, &p)
	}
	return prfs, nil
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

// CreateCustomTokenWithClaims returns token as a string
func (db *Firebase) CreateCustomTokenWithClaims(ctx context.Context, uid string, claims map[string]interface{}) (string, error) {
	return db.Auth.CustomTokenWithClaims(ctx, uid, claims)
}

// VerifyToken verify JWT handled by middleware.go returning the uid
func (db *Firebase) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	t, err := db.Auth.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// IsAdminClaims returns token as a string. ! Not Used in admin middleware.go.
// ! currently not in use because of method below
func (db *Firebase) IsAdminClaims(claims map[string]interface{}) bool {
	// The claims can be accessed on the user record.
	log.Infoln(claims)
	if admin, ok := claims["is_admin"]; ok {
		if admin.(bool) {
			return admin.(bool)
		}
	}
	return false
}

// IsAdminUID will return true if the uid is found in the admin fb storage
// It's being called in loginCreateToken handler
func (db *Firebase) IsAdminUID(ctx context.Context, uid string) (bool, error) {
	path := os.Getenv("ENV") + "/admins"
	ref := db.Client.NewRef(path)
	var isAdmin map[string]float64
	if err := ref.Get(ctx, &isAdmin); err != nil {
		return false, err
	}
	if _, ok := isAdmin[uid]; ok {
		return true, nil
	}
	return false, nil
}

// IsAdminUID will return true if the uid is found in the admin fb storage
// It's being called in loginCreateToken handler
func (db *Firebase) IsProfessional(ctx context.Context, uid string) (isPro bool, err error) {
	var profile storage.FirebaseProfile
	path := os.Getenv("ENV") + "/profiles"
	ref := db.Client.NewRef(path).Child(uid)
	if err := ref.Get(ctx, &profile); err != nil {
		return false, err
	}
	if !profile.IsProfessional {
		return false, errors.Errorf("User %s is not a professional", profile.DisplayName)
	}
	return true, nil
}
