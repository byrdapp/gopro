package storage

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/byblix/gopro/storage"

	// Postgres driver
	_ "github.com/lib/pq"
)

type postgres struct {
	DB *sql.DB
}

// NewPQ Starts ORM
func NewPQ() (storage.Service, error) {
	fmt.Println("Starting the PGSQL DB")

	connStr := os.Getenv("POSTGRES_CONNSTR")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	logrus.Infoln("Starting postgres DB")
	return &postgres{db}, nil
}

func (p *postgres) Save(str string) (string, error) {
	return "", nil
}

func (p *postgres) Close() error {
	err := p.DB.Close()
	if err != nil {
		return err
	}
	return nil
}
