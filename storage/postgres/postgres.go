package storage

import (
	"context"
	"database/sql"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"

	qb "github.com/Masterminds/squirrel"

	"github.com/blixenkrone/gopro/utils/logger"

	// Postgres driver
	_ "github.com/lib/pq"
)

// Postgres is the database
type Postgres struct {
	DB *sql.DB
}

var log = logger.NewLogger()

// NewPQ Starts ORM
func NewPQ() (Service, error) {
	log.Info("Starting postgres...")
	connStr, ok := os.LookupEnv("POSTGRES_CONNSTR")
	if !ok {
		log.Fatal("Error opening postgress connstr")
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Infoln("Started psql DB")
	return &Postgres{db}, nil
}

// CreateMedia -
func (p *Postgres) CreateMedia(ctx context.Context, media *Media) (string, error) {
	var id int64
	err := p.DB.QueryRowContext(ctx, "INSERT INTO media(name, user_id, display_name) VALUES($1, $2, $3) RETURNING id;", media.Name, media.UserID, media.DisplayName).Scan(&id)
	if err != nil {
		p.HandleRowError(err)
		return "", err
	}
	log.Infof("Inserted new media with id: %v", id)
	return strconv.Itoa(int(id)), nil
}

// GetMediaByID -
func (p *Postgres) GetMediaByID(ctx context.Context, id string) (*Media, error) {
	var media Media
	sqlid, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	row := p.DB.QueryRowContext(ctx, `SELECT * FROM media WHERE id = $1`, sqlid)
	err = row.Scan(&media.ID, &media.Name, &media.UserID, &media.DisplayName)
	if err != nil {
		p.HandleRowError(err)
		return nil, err
	}
	return &media, nil
}

// GetMedias -
func (p *Postgres) GetMedias(ctx context.Context, params ...[]string) ([]*Media, error) {
	var medias []*Media
	rows, err := p.DB.QueryContext(ctx, "SELECT * FROM media LIMIT 10;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var media Media
		if err := rows.Scan(&media.ID, &media.Name, &media.UserID, &media.DisplayName); err != nil {
			return nil, err
		}
		medias = append(medias, &media)
	}

	if err := p.CancelRowsError(rows); err != nil {
		log.Errorf("Error getting rows: %s", err)
		return nil, err
	}

	return medias, nil
}

// CreateProfessional -
func (p *Postgres) CreateProfessional(ctx context.Context, pro *Professional) (string, error) {
	var id int64
	err := p.DB.QueryRowContext(ctx, "INSERT INTO professional(name, user_id, display_name, email) VALUES($1, $2, $3, $4) RETURNING id;", pro.Name, pro.UserID, pro.DisplayName, pro.Email).Scan(&id)
	if err != nil {
		p.HandleRowError(err)
		return "", err
	}
	log.Infof("Inserted new pro with id: %v", id)
	return strconv.Itoa(int(id)), nil
}

// GetProProfile -
func (p *Postgres) GetProProfile(ctx context.Context, id string) (*Professional, error) {
	var pro Professional
	// query, _, err := qb.Select("*").From("professional").Where("id", id).ToSql()
	// if err != nil {
	// 	return nil, err
	// }
	// log.Infoln(query)
	query := "SELECT * FROM professional WHERE id = $1"
	row := p.DB.QueryRowContext(ctx, query, id)
	if err := row.Scan(&pro.ID, &pro.DisplayName, &pro.UserID, &pro.Name, &pro.Email, &pro.StatsID); err != nil {
		return nil, err
	}
	return &pro, nil
}

// GetProProfileByEmail -
func (p *Postgres) GetProProfileByEmail(ctx context.Context, email string) (*Professional, error) {
	var pro Professional
	query, i, err := qb.Select("*").From("professional").Where("email = ?", email).ToSql()
	if err != nil {
		return nil, err
	}
	spew.Dump(i)
	log.Infoln(query)
	// query := "SELECT * FROM professional WHERE id = $1"
	row := p.DB.QueryRowContext(ctx, query)
	if err := row.Scan(&pro.ID, &pro.DisplayName, &pro.UserID, &pro.Name, &pro.Email, &pro.StatsID); err != nil {
		return nil, err
	}
	return &pro, nil
}

// CreateProStats adds stats from the pro ID column to stats table
func (p *Postgres) CreateProStats(ctx context.Context, stats *Stats) (int64, error) {
	var id int64
	err := p.DB.QueryRowContext(ctx,
		"INSERT INTO stats(accepted_assignments, device, sales_amount, sales_quantity) VALUES ($2, $3, $4, $5) WHERE id = $1 RETURNING id;",
		stats.AcceptedAssignments, stats.Device, stats.SalesAmount, stats.SalesQuantity).Scan(&id)
	if err != nil {
		return -1, err
	}
	log.Infof("Added pro stats with id: %v", id)
	return id, nil
}

// GetProStats returns stats given an ID for a professional
func (p *Postgres) GetProStats(ctx context.Context, proID string) (*Stats, error) {
	var stats Stats
	query := "SELECT * FROM stats WHERE id = $1"
	row := p.DB.QueryRowContext(ctx, query, proID)
	err := row.Scan(&stats.ID, &stats.AcceptedAssignments, &stats.Device, &stats.SalesAmount, &stats.SalesQuantity)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetProByID -
func (p *Postgres) GetProByID(ctx context.Context, id string) (*Professional, error) {
	var pro Professional
	sqlid, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	row := p.DB.QueryRowContext(ctx, `SELECT * FROM media WHERE id = $1`, sqlid)
	err = row.Scan(&pro.ID, &pro.Name, &pro.UserID, &pro.DisplayName, &pro.Email)
	if err != nil {
		p.HandleRowError(err)
		return nil, err
	}
	return &pro, nil
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
		log.Errorf("No rows were returned: %s\n", err)
	case err != nil:
		log.Errorf("Error with query: %v\n", err)
	default:
		log.Panicf("Default error: %v\n", err)
	}
}

// CancelRowsError to handle errors from sql requests
func (p *Postgres) CancelRowsError(rows *sql.Rows) error {
	if err := rows.Close(); err != nil {
		return err
	}
	return rows.Err()
}
