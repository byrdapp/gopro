package old_postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/Masterminds/squirrel"

	"github.com/blixenkrone/gopro/internal/storage"
	"github.com/blixenkrone/gopro/pkg/logger"
	_ "github.com/lib/pq"
)

var (
	qb  = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	log = logger.NewLogger()
)

// Old_Postgres is the database
type Old_Postgres struct {
	DB *sql.DB
}

// NewPQ Starts ORM
// func NewPQ(s storage.Service) (*Postgres, error) {
func NewPQ() (storage.PQService, error) {
	connStr, ok := os.LookupEnv("POSTGRES_CONNSTR")
	if !ok {
		return nil, errors.New("Error opening postgress connstr from environment variable")
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
	return &Old_Postgres{db}, nil
}

/** BOOKING ENDPOINTS */

// CreateBooking is being made from the media client
func (p *Old_Postgres) CreateBooking(ctx context.Context, proUID string, b storage.Booking) (bookingID string, err error) {
	sb := qb.RunWith(p.DB)
	err = sb.Insert("booking").Columns(
		"user_uid", "media_uid", "task", "price", "credits", "date_start", "date_end", "lat", "lng").Values(
		proUID, &b.MediaUID, &b.MediaBooker, &b.Task, &b.Price, &b.Credits, &b.DateStart, &b.DateEnd, &b.Lat, &b.Lng,
	).Suffix("RETURNING id").QueryRowContext(ctx).Scan(&bookingID)
	if err != nil {
		log.Errorf("Insert error: %s", err)
		return "", err
	}
	return bookingID, nil
}

// GetBookings gets all the bookings from a professional user by ID
func (p *Old_Postgres) GetBookingsByUID(ctx context.Context, proID string) ([]*storage.Booking, error) {
	var bookings []*storage.Booking
	sb := qb.RunWith(p.DB)
	rows, err := sb.Select("*").From("booking").Where("user_uid = ?", proID).OrderBy("created_at DESC").QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b storage.Booking
		err := rows.Scan(&b.ID, &b.UserUID, &b.MediaUID, &b.MediaBooker, &b.Task, &b.Price, &b.Credits, &b.IsActive, &b.IsCompleted, &b.DateStart, &b.DateEnd, &b.CreatedAt, &b.Lat, &b.Lng)
		if err != nil {
			return nil, err
		}
		if err := p.HandleRowError(err); err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}
	return bookings, nil
}

// UpdateBooking -
func (p *Old_Postgres) UpdateBooking(ctx context.Context, b *storage.Booking) error {
	sb := qb.RunWith(p.DB)
	_, err := sb.Update("booking").
		Set("is_active", &b.IsActive).
		Set("is_completed", &b.IsCompleted).
		Set("task", &b.Task).
		Where("id = ?", &b.ID).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// DeleteBooking -
func (p *Old_Postgres) DeleteBooking(ctx context.Context, bookingID string) error {
	sb := qb.RunWith(p.DB)
	_, err := sb.Delete("booking").
		Where("id = ?", bookingID).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetBookingsAdmin returns bookings sorted by created_at date with crossjoined profile uid's.
func (p *Old_Postgres) GetBookingsAdmin(ctx context.Context) (res []*storage.AdminBookings, err error) {
	query, _, err := qb.Select("booking.task", "booking.credits", "booking.price", "booking.created_at", "booking.is_active", "professional.id", "professional.user_uid", "professional.pro_level").
		From("booking").
		LeftJoin("professional ON booking.user_uid = professional.user_uid").
		OrderBy("booking.created_at DESC", "booking.is_active DESC").
		Limit(5).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var j storage.AdminBookings
		if err := rows.Scan(&j.Booking.Task, &j.Booking.Credits, &j.Booking.Price, &j.Booking.CreatedAt, &j.Booking.IsActive, &j.Professional.ID, &j.Professional.UserUID, &j.Professional.ProLevel); err != nil {
			return nil, err
		}
		if err := p.HandleRowError(rows.Err()); err != nil {
			return nil, err
		}
		res = append(res, &j)
	}
	return res, nil
}

/**
 * PROFESSIONAL
 */

// GetProfile -
func (p *Old_Postgres) GetProfile(ctx context.Context, id string) (*storage.Professional, error) {
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

/**
 * DIV
 */

// Ping to see if theres connection
func (p *Old_Postgres) Ping() error {
	return p.DB.Ping()
}

// Close -
func (p *Old_Postgres) Close() error {
	err := p.DB.Close()
	if err != nil {
		return err
	}
	return nil
}

// HandleRowError to handle errors from sql requests
func (p *Old_Postgres) HandleRowError(err error) error {
	switch err {
	case sql.ErrNoRows:
		return sql.ErrNoRows
	case nil:
		return err
	default:
		return err
	}
}

// CancelRowsError to handle errors from sql requests
func (p *Old_Postgres) CancelRowsError(rows *sql.Rows) error {
	if err := rows.Close(); err != nil {
		return err
	}
	return rows.Err()
}
