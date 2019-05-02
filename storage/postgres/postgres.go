package storage

import (
	"context"
	"database/sql"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	// Postgres driver
	_ "github.com/lib/pq"
)

// Postgres is the database
type Postgres struct {
	DB *sql.DB
}

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
	var id int64
	ctx := context.Background()
	err := p.DB.QueryRowContext(ctx, "INSERT INTO media(name, user_id, display_name) VALUES($1, $2, $3) RETURNING id;", media.Name, media.UserID, media.DisplayName).Scan(&id)
	if err != nil {
		p.HandleRowError(err)
		return "", err
	}
	logrus.Infof("Inserted new media with id: %v", id)
	return strconv.Itoa(int(id)), nil
}

// GetMediaByID -
func (p *Postgres) GetMediaByID(id string) (*Media, error) {
	var media Media
	sqlid, _ := strconv.Atoi(id)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	row := p.DB.QueryRowContext(ctx, `SELECT * FROM media WHERE id = $1`, sqlid)
	err := row.Scan(&media.ID, &media.Name, &media.UserID, &media.DisplayName)
	if err != nil {
		p.HandleRowError(err)
		return nil, err
	}
	return &media, nil
}

// GetMedias -
func (p *Postgres) GetMedias(params ...[]string) ([]*Media, error) {
	medias := make([]*Media, 0)
	ctx := context.Background()
	rows, err := p.DB.QueryContext(ctx, "SELECT * FROM media;", params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var media *Media
		if err := rows.Scan(&media); err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}

	if err := p.CancelRowsError(rows); err != nil {
		return nil, err
	}

	return medias, nil
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

// HandleRowError to handle errors from sql requests
func (p *Postgres) HandleRowError(err error) {
	switch {
	case err == sql.ErrNoRows:
		logrus.Errorf("No rows were returned: %s\n", err)
	case err != nil:
		logrus.Errorf("Error with query: %v\n", err)
	default:
		logrus.Panicf("Default error: %v\n", err)
	}
}

// CancelRowsError to handle errors from sql requests
func (p *Postgres) CancelRowsError(rows *sql.Rows) error {
	if err := rows.Close(); err != nil {
		return err
	}
	return rows.Err()
}
