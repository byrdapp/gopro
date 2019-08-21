package postgres

import (
	"context"
	"database/sql"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"

	squirrel "github.com/Masterminds/squirrel"

	"github.com/blixenkrone/gopro/storage"
	"github.com/blixenkrone/gopro/utils/logger"

	// Postgres driver
	_ "github.com/lib/pq"
)

var (
	qb  = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	log = logger.NewLogger()
)

// Postgres is the database
type Postgres struct {
	DB *sql.DB
}

// NewPQ Starts ORM
func NewPQ() (storage.PQService, error) {
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

/**
 * BOOKING
 */

// CreateBooking is being made from the media client
func (p *Postgres) CreateBooking(ctx context.Context, proUID string, b storage.Booking) (bookingID string, err error) {
	sb := qb.RunWith(p.DB)
	err = sb.Insert("booking").Columns(
		"user_uid_fk", "media_uid", "media_booker", "task", "price", "credits", "date_start", "date_end", "lat", "lng").Values(
		proUID, &b.MediaUID, &b.MediaBooker, &b.Task, &b.Price, &b.Credits, &b.DateStart, &b.DateEnd, &b.Lat, &b.Lng,
	).Suffix("RETURNING id").QueryRowContext(ctx).Scan(&bookingID)
	if err != nil {
		log.Errorf("Insert error: %s", err)
		return "", err
	}
	return bookingID, nil
}

// GetProBookings gets all the bookings from a professional user by ID
func (p *Postgres) GetProBookings(ctx context.Context, proID string) ([]*storage.Booking, error) {
	var b storage.Booking
	query, i, err := qb.Select("*").From("booking").Where("pro_id = ?", proID).ToSql()
	if err != nil {
		return nil, err
	}
	log.Infoln(i)
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	for {
		if rows.Next() {
			err := rows.Scan(&b.ID)
			if err != nil {
				return nil, err
			}
		}
		err := rows.Err()
		if err != nil {
			return nil, err
		}
	}
}

/**
 * PROFESSIONAL
 */

// CreateProfessional under construction
func (p *Postgres) CreateProfessional(ctx context.Context, pro *storage.Professional) (string, error) {
	var id int64
	// err := p.DB.QueryRowContext(ctx, "INSERT INTO professional(name, user_id, display_name, email) VALUES($1, $2, $3, $4) RETURNING id;", pro.Name, pro.UserID, pro.DisplayName, pro.Email).Scan(&id)
	// if err != nil {
	// 	p.HandleRowError(err)
	// 	return "", err
	// }
	log.Infof("Inserted new pro with id: %v", id)
	return strconv.Itoa(int(id)), nil
}

// GetProProfile -
func (p *Postgres) GetProProfile(ctx context.Context, id string) (*storage.Professional, error) {
	var pro storage.Professional
	// query, _, err := qb.Select("*").From("professional").Where("id", id).ToSql()
	// if err != nil {
	// 	return nil, err
	// }
	// log.Infoln(query)
	query := "SELECT * FROM professional WHERE id = $1"
	row := p.DB.QueryRowContext(ctx, query, id)
	if err := row.Scan(&pro.ID); err != nil {
		return nil, err
	}
	return &pro, nil
}

// GetProProfileByEmail -
func (p *Postgres) GetProProfileByEmail(ctx context.Context, email string) (*storage.Professional, error) {
	var pro storage.Professional
	query, i, err := qb.Select("*").From("professional").Where("email = ?", email).ToSql()
	if err != nil {
		return nil, err
	}
	spew.Dump(i)
	log.Infoln(query)
	// query := "SELECT * FROM professional WHERE id = $1"
	row := p.DB.QueryRowContext(ctx, query)
	if err := row.Scan(&pro.ID, &pro); err != nil {
		return nil, err
	}
	return &pro, nil
}

// GetProByID -
func (p *Postgres) GetProByID(ctx context.Context, id string) (*storage.Professional, error) {
	var pro storage.Professional
	sqlid, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	row := p.DB.QueryRowContext(ctx, `SELECT * FROM media WHERE id = $1`, sqlid)
	err = row.Scan(&pro.ID)
	if err != nil {
		p.HandleRowError(err)
		return nil, err
	}
	return &pro, nil
}

/**
 * DIV
 */

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
