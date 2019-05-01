package storage

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	// Postgres driver
	_ "github.com/lib/pq"
)

// Postgres is the database
type Postgres struct {
	DB *sql.DB
}

var ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)

// NewPQ Starts ORM
func NewPQ() (Service, error) {
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
func (p *Postgres) AddMedia(media *Media) error {
	defer cancel()
	p.DB.QueryRowContext(ctx, "INSERT INTO media ").Scan(media)
	// p.DB.ExecContext()
	return nil
}

// GetMediaByID -
func (p *Postgres) GetMediaByID(id string) (*Media, error) {
	sqlid, _ := strconv.Atoi(id)
	logrus.Infof("ID is = %v", sqlid)
	var media Media
	row := p.DB.QueryRowContext(ctx, `SELECT * FROM media WHERE id = ?`, sqlid)
	err := row.Scan(&media.ID, &media.ProfileID, &media.DisplayName, &media.Address)
	if err != nil {
		p.HandleError(err)
		return nil, err
	}
	return &media, nil
}

// GetMedias -
func (p *Postgres) GetMedias() ([]*Media, error) {
	// var medias []*media
	return nil, nil
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

// HandleError to handle errors from sql requests
func (p *Postgres) HandleError(err error) {
	switch err {
	case sql.ErrNoRows:
		logrus.Errorf("No rows were returned: %s\n", err)
	case nil:
		logrus.Errorf("Error with query: %v\n", err)
	default:
		logrus.Panicf("Some error: %v\n", err)
	}
}
