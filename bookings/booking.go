package bookings

import (
	"time"

	"github.com/blixenkrone/gopro/models"
	"github.com/sirupsen/logrus"
)

type timeStamps struct {
	TimeBegin *time.Time // .now()
	TimeEnd   *time.Time
}

type geoLocation struct {
	Lat, Lng float64
}

// Booking -
type Booking struct {
	Profile *models.ProfileProps
	Task    string
	Area    *geoLocation
	Price   int
	Time    *timeStamps
	log     *logrus.Logger
}

// BookingService -
type BookingService interface {
	CreateBooking() error
	DeleteBooking() error
	GetBooking() error
	UpdateBooking() error
	MarkBookingAsDone() error
}

/**
 * interface{} with functions needed?
 */

// New creates a new Booking
func New(profile *models.ProfileProps, task string) (BookingService, error) {
	return &Booking{
		Profile: profile,
		Task:    task,
	}, nil
}

func (b *Booking) CreateBooking() error     { return nil }
func (b *Booking) DeleteBooking() error     { return nil }
func (b *Booking) GetBooking() error        { return nil }
func (b *Booking) UpdateBooking() error     { return nil }
func (b *Booking) MarkBookingAsDone() error { return nil }
