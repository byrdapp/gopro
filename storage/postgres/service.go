package storage

// Service is storage service interface that exports CRUD data from CLIENT -> API -> postgres db via http
type Service interface {
	Save(string) (string, error)
	// Load(string) (string, error)
	AddMedia(*Media) error
	// Delete(string) (string, error)
	Close() error
	GetMediaByID(string) (*Media, error)
	GetMedias() ([]*Media, error)
	Ping() error
	HandleError(error)
}

// Media is for SQL media queries
type Media struct {
	ID          int    `json:"id"`
	ProfileID   string `json:"profile_id"`
	DisplayName string `json:"display_name"`
	Address     string `json:"address"`
}
