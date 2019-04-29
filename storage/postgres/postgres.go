package storage

import (
	"context"
	"database/sql"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/byblix/gopro/storage"

	// Postgres driver
	_ "github.com/lib/pq"
)

// Service.Postgres is the database
type Postgres struct {
	DB *sql.DB
}

var ctx, cancel = context.WithCancel(context.Background())

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
func (p *Postgres) AddMedia() {
	p.DB.QueryContext(ctx, `INSERT INTO media()`)
}

// GetMediaByID -
func (p *Postgres) GetMediaByID(id string) error {
	logrus.Info("Get the media by ID")
	// p.DB.QueryContext(ctx, `INSERT INTO media()`)
	return nil
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
