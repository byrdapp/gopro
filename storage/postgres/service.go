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
	Ping() error
	HandleRowError(error)
	CancelRowsError(*sql.Rows) error
}

// Media is for SQL media queries
type Media struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	// CreatedDate time.Time `json:"createdDate"`
}
