package storage

import (
	"context"
	"database/sql"
)

// Service is storage service interface that exports CRUD data from CLIENT -> API -> postgres db via http
type Service interface {
	UpdateMedia(id string) (*Media, error)
	// Load(string) (string, error)
	CreateMedia(context.Context, *Media) (string, error)
	// Delete(string) (string, error)
	Close() error
	GetMediaByID(context.Context, string) (*Media, error)
	GetMedias(context.Context, ...[]string) ([]*Media, error)
	GetProProfile(ctx context.Context, id string) (*Professional, error)
	CreateProfessional(context.Context, *Professional) (string, error)
	GetProStats(ctx context.Context, proID string) (*Stats, error)
	CreateProStats(context.Context, *Stats) (int64, error)
	Ping() error
	HandleRowError(error)
	CancelRowsError(*sql.Rows) error
}

// Media is for SQL media queries
type Media struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	UserID      string `json:"userId,sql:userId"`
	DisplayName string `json:"displayName,sql:display_name"`
	// Email       string `json:"email"`
	// CreatedDate time.Time `json:"createdDate"`
}

// Professional user class
type Professional struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	UserID      string `json:"userId,sql:user_id"`
	DisplayName string `json:"displayName,sql:display_name"`
	Email       string `json:"email"`
	StatsID     int    `json:"statsID,sql:stats_id"`
}

// Stats from a professional
type Stats struct {
	ID                  int    `json:"id"`
	SalesQuantity       int64  `json:"salesQuantity,sales_quantity,omitempty"`
	SalesAmount         int64  `json:"salesAmount,sales_amount,omitempty"`
	AcceptedAssignments int    `json:"acceptedAssignments,sql:accepted_assignments,omitempty"`
	Device              string `json:"deviceBrand,sql:device,omitempty"`
}
