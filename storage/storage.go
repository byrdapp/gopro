package storage

import (
	"context"
	"database/sql"
	"time"

	"firebase.google.com/go/auth"
)

// PQService is storage service interface that exports CRUD data from CLIENT -> API -> postgres db via http
type PQService interface {
	GetBookings(ctx context.Context, proID string) ([]*Booking, error)
	CreateBooking(ctx context.Context, uid string, b Booking) (string, error)
	UpdateBooking(ctx context.Context, b *Booking) error
	DeleteBooking(ctx context.Context, bookingID string) error
	GetBookingsAdmin(ctx context.Context) ([]*BookingProfessional, error)
	GetProfile(ctx context.Context, id string) (*Professional, error)
	Close() error
	Ping() error
	HandleRowError(error) error
	CancelRowsError(*sql.Rows) error
}

// FBService contains firebase methods
type FBService interface {
	GetTransactions() ([]*Transaction, error)
	// ! dont update anything from the API - only scripting
	UpdateData(uid string, prop string, value string) error
	GetWithdrawals() ([]*Withdrawals, error)
	GetProfile(ctx context.Context, uid string) (*FirebaseProfile, error)
	GetProfileByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	GetProfiles(ctx context.Context) ([]*FirebaseProfile, error)
	GetAuth() ([]*auth.ExportedUserRecord, error)
	DeleteAuthUserByUID(uid string) error
	CreateCustomTokenWithClaims(ctx context.Context, uid string, claims map[string]interface{}) (string, error)
	IsAdminClaims(claims map[string]interface{}) bool
	IsAdminUID(ctx context.Context, uid string) (bool, error)
	VerifyToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// Professional user class
type Professional struct {
	ID       string `json:"id" sql:"id"`
	UserUID  string `json:"userUID" sql:"user_uid"`
	ProLevel int    `json:"proLevel" sql:"pro_level"`
}

// Booking repræsents a professional user appointment from a media
type Booking struct {
	ID          string     `json:"id,omitempty" sql:"id"`
	MediaUID    string     `json:"mediaUID,omitempty" sql:"media_uid"`
	MediaBooker string     `json:"mediaBooker,omitempty" sql:"media_booker"`
	UserUID     string     `json:"userUID,omitempty" sql:"user_uid"`
	Task        string     `json:"task,omitempty"`
	Price       int        `json:"price,omitempty"`
	Credits     int        `json:"credits,omitempty"`
	IsActive    bool       `json:"isActive,omitempty" sql:"is_active"`
	IsCompleted bool       `json:"isCompleted,omitempty" sql:"is_completed"`
	DateStart   *time.Time `json:"dateStart,omitempty" sql:"date_start"`
	DateEnd     *time.Time `json:"dateEnd,omitempty" sql:"date_end"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" sql:"created_at"`
	Lng         string     `json:"lng,omitempty" sql:"lng"`
	Lat         string     `json:"lat,omitempty" sql:"lat"`
}

// BookingProfessional is a joined response for a booking attached to a pro user
type BookingProfessional struct {
	Booking
	Professional
	FirebaseProfile
}

// FirebaseProfile defines a profile in firebsse
type FirebaseProfile struct {
	UserID              string `json:"userId,omitempty"`
	DisplayName         string `json:"displayName"`
	FirstName           string `json:"firstName,omitempty"`
	LastName            string `json:"lastName,omitempty"`
	Address             string `json:"address,omitempty"`
	Email               string `json:"email,omitempty"`
	IsMedia             bool   `json:"isMedia,omitempty"`
	IsProfessional      bool   `json:"isProfessional,omitempty"`
	IsPress             bool   `json:"isPress,omitempty"`
	SalesQuantity       int64  `json:"salesQuantity,omitempty"`
	SalesAmount         int64  `json:"salesAmount,omitempty"`
	WithdrawableAmount  int64  `json:"withdrawableAmount,omitempty"`
	AcceptedAssignments int    `json:"acceptedAssignments,omitempty"`
	UserPicture         string `json:"userPicture,omitempty"`
	SoldStories         int    `json:"soldStories,omitempty"`
	DeviceBrand         string `json:"deviceBrand,omitempty"`
	DeviceModel         string `json:"deviceModel,omitempty"`
	OsSystem            string `json:"osSystem,omitempty"`
	UploadedStories     int    `json:"uploadedStories,omitempty"`
}

// CreateDate         *time.Time `json:"createDate"`

// Media struct
type Media struct {
	ID                  string `sql:"id"`
	ProfileData         *FirebaseProfile
	UserID              string `json:"userId"`
	Country             string `json:"country,omitempty"`
	City                string `json:"city,omitempty"`
	GoCredits           byte   `json:"goCredits,omitempty"`
	SubscriptionCredits byte   `json:"subscriptionCredits,omitempty"`
}

// Transaction struct
type Transaction struct {
	PaymentDate              uint64 `json:"paymentDate,omitempty"`
	StoryID                  string `json:"storyId,omitempty"`
	PaymentSellerDisplayName string `json:"paymentSellerDisplayName,omitempty"`
	PaymentSeller            string `json:"paymentSeller,omitempty"`
	PaymentBuyer             string `json:"paymentBuyer,omitempty"`
	PaymentBuyerDisplayName  string `json:"paymentBuyerDisplayName,omitempty"`
}

// Withdrawals ..
type Withdrawals struct {
	// CashoutDetails       string `json:"cashoutDetails,omitempty"`
	RequestAmount        int64  `json:"requestAmount,omitempty"`
	RequestCompleted     bool   `json:"requestCompleted,omitempty"`
	RequestCompletedDate int64  `json:"requestCompletedDate,omitempty"`
	RequestUserID        string `json:"requestUser,omitempty"`
	RequestDate          int64  `json:"requestDate,omitempty"`
}
