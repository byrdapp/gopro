package errors

import (
	"encoding/json"
	"net/http"
	"sync"

	logger "github.com/blixenkrone/gopro/utils/logger"
	fmterr "github.com/pkg/errors"
)

// ErrorBuilder builds custom errors
type ErrorBuilder struct {
	Code      int    `json:"code"`
	ClientMsg string `json:"msg"`
	w         http.ResponseWriter
	err       error
	s         sync.Once
}

var log = logger.NewLogger()

// NewResErr -
func NewResErr(err error, msg string, code int, w http.ResponseWriter) {
	build := &ErrorBuilder{
		Code:      code,
		ClientMsg: msg,
		w:         w,
		err:       err,
	}
	w.WriteHeader(code)
	build.errResponseLogger()
}

func (e *ErrorBuilder) errStackTraced() {
	switch err := fmterr.Cause(e.err).(type) {
	// case err:
	// ! build error handler stack trace

	// handle specifically
	default:
		// unknown error
	}
	newErr := fmterr.Wrap(e.err, e.ClientMsg)
	log.Errorf("Original error: %s\n", e.err)
	log.Errorf("New error: %s\n", newErr)
}

// ErrResponseLogger defines what error goes to the log and what to display as JSON in client
func (e *ErrorBuilder) errResponseLogger() {
	jsonParseErr := json.NewEncoder(e.w).Encode(e)
	if jsonParseErr != nil {
		log.Errorf("Json parse error: %s\n", jsonParseErr)
	}
}
