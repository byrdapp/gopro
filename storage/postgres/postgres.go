package storage

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/byblix/gopro/storage"

	// Postgres driver
	_ "github.com/lib/pq"
)

// Service.Postgres is the database
type Postgres struct {
	DB *sql.DB
}

var ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)

// NewPQ Starts ORM
func NewPQ() (storage.Service, error) {
	connStr := os.Getenv("POSTGRES_CONNSTR")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	logrus.Infoln("Started psql DB")
	return &Postgres{db}, nil
}

// Save -
func (p *Postgres) Save(str string) (string, error) {
	return "", nil
}

// AddMedia -
func (p *Postgres) AddMedia(media *storage.Media) error {
	defer cancel()
	p.DB.QueryRowContext(ctx, "INSERT INTO media ").Scan(media)
	// p.DB.ExecContext()
	return nil
}

// GetMediaByID -
func (p *Postgres) GetMediaByID(id string) (*storage.Media, error) {
	logrus.Info(id)
	var media storage.Media
	// sqlStatement := `SELECT id, display_name FROM media WHERE id = $1`
	row := p.DB.QueryRow("SELECT * FROM media")
	err := row.Scan(&media.UserID, &media.ProfileData.DisplayName, &media.ID)
	if err != nil {
		handleError(err, row)
		return nil, err
	}
	return &media, nil
}

// Ping to see if theres connection
func (p *Postgres) Ping() error {
	return p.Ping()
}

// Close -
func (p *Postgres) Close() error {
	err := p.DB.Close()
	if err != nil {
		return err
	}
	return nil
}

func handleError(err error, data interface{}) {
	switch err {
	case sql.ErrNoRows:
		logrus.Println("No rows were returned")
	case nil:
		logrus.Printf("query error: %v\n", err)
	default:
		logrus.Panicf("Some error: %v\n", err)
	}
}
