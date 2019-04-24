package storage

// Service is storage service interface that exports CRUD data from CLIENT -> API -> postgres db via http
type Service interface {
	Save(string) (string, error)
	// Load(string) (string, error)
	AddMedia()
	// Delete(string) (string, error)
	Close() error
}

type media struct {
	ID        int    `json:"id"`
	ProfileID string `json:"profile_id"`
}
