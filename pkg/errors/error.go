package errors

import (
	"encoding/json"
	"net/http"

	logger "github.com/blixenkrone/gopro/pkg/logger"
	"github.com/pkg/errors"
)

var log = logger.NewLogger()

// ErrorBuilder builds custom errors to a http response writer
type ErrorBuilder struct {
	Code      int    `json:"code"`
	ClientMsg string `json:"msg"`
	w         http.ResponseWriter
	err       error
	traced    string
}

// NewResErr constructs and executes an err struct.
// Set stackTraced = "trace" to show error stack.
func NewResErr(err error, msg string, code int, w http.ResponseWriter, stackTraced ...string) {
	build := &ErrorBuilder{
		Code:      code,
		ClientMsg: msg,
		w:         w,
		err:       err,
	}
	if len(stackTraced) > 0 {
		build.traced = stackTraced[0]
	}

	w.WriteHeader(code)
	if build.traced == "trace" {
		build.errStackTraced()
	}
	build.errResponseLogger()
}

func (e *ErrorBuilder) errStackTraced() {
	err := errors.WithStack(e.err)
	log.Errorf("Originated error: %+v", err)
}

// ErrResponseLogger defines what error goes to the log and what to display as JSON in client
func (e *ErrorBuilder) errResponseLogger() {
	jsonParseErr := json.NewEncoder(e.w).Encode(e)
	if jsonParseErr != nil {
		log.Errorf("Json parse error: %s\n", jsonParseErr)
	}
}
