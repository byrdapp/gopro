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

func (p *Postgres) Save(str string) (string, error) {
	return "", nil
}

func (p *Postgres) AddMedia() {
	p.DB.QueryContext(ctx, `INSERT INTO media()`)
}

func (p *Postgres) GetMediaByID(id string) {

	p.DB.QueryContext(ctx, `INSERT INTO media()`)
}

func (p *Postgres) Close() error {
	err := p.DB.Close()
	if err != nil {
		return err
	}
	return nil
}
