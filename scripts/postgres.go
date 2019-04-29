package scripts

import (
	psqr "github.com/byblix/gopro/storage/postgres"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/byblix/gopro/storage"
)

// ExportToPostgres -
func ExportToPostgres(w http.ResponseWriter, r *http.Request) {
	fbdb, err := storage.InitFirebaseDB()
	if err != nil {
		logrus.Errorf("Err: %s", err)
	}
	pgdb, err := psqr.NewPQ()
	if err != nil {
		logrus.Errorf("Error POSQ: %s", err)
	}
	if err := pgdb.Ping(); err != nil {
		logrus.Errorf("Err %s", err)
	}

	prfs, err := fbdb.GetProfiles()
	if err != nil {
		logrus.Errorf("Err: %s", err)
	}
	for _, val := range prfs[:3] {
		logrus.Info(val.UserID)
	}
}
