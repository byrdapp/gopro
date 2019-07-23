package errors

import (
	"encoding/json"
	"net/http"

	logger "github.com/blixenkrone/gopro/utils/logger"
)

// ErrorBuilder builds custom errors
type ErrorBuilder struct {
	Code      int    `json:"code"`
	ClientMsg string `json:"msg"`
}

var log = logger.NewLogger()

// ErrResponseLogger defines what error goes to the log and what to display as JSON in client
func (resErr *ErrorBuilder) ErrResponseLogger(err error, w http.ResponseWriter) {
	log.Errorf("Error: %s. Code: %v", resErr.ClientMsg, resErr.Code)
	jsonParseErr := json.NewEncoder(w).Encode(resErr)
	if jsonParseErr != nil {
		log.Errorf("Json parse error: %s", err)
	}
}
