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

// UpdateMedia -
func (p *Postgres) UpdateMedia(id string) (*Media, error) {
	// todo: also alters the departments or new prototype?
	var media Media
	return &media, nil
}

// CreateMedia -
func (p *Postgres) CreateMedia(media *Media) (string, error) {
	defer cancel()
	var id int64
	err := p.DB.QueryRowContext(ctx, "INSERT INTO media(name, user_id, display_name) VALUES($1, $2, $3) RETURNING id;", media.Name, media.UserID, media.DisplayName).Scan(&id)
	if err != nil {
		p.HandleError(err)
		return "", err
	}
	logrus.Infof("Inserted new media with id: %v", id)
	return strconv.Itoa(int(id)), nil
}

// GetMediaByID -
func (p *Postgres) GetMediaByID(id string) (*Media, error) {
	var media Media
	sqlid, _ := strconv.Atoi(id)
	ctx, cancel = context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	row := p.DB.QueryRow(`SELECT * FROM media WHERE id = $1`, sqlid)
	err := row.Scan(&media.ID, &media.Name, &media.UserID, &media.DisplayName)
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
	return p.DB.Ping()
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
		logrus.Panicf("Default error: %v\n", err)
	}
}
