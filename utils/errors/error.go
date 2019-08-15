package errors

import (
	"encoding/json"
	"net/http"
	"sync"

	logger "github.com/blixenkrone/gopro/utils/logger"
	"github.com/pkg/errors"
)

// ErrorBuilder builds custom errors
type ErrorBuilder struct {
	Code      int    `json:"code"`
	ClientMsg string `json:"msg"`
	w         http.ResponseWriter
	err       error
	s         sync.Once
}

type stackTracer interface {
	StackTrace() errors.StackTrace
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
	build.errStackTraced()
}

func (e *ErrorBuilder) errStackTraced() {
	// cause := errors.New("Error stacktrace: ")
	err := errors.WithStack(e.err)
	log.Errorf("%+v", err)
}

// ErrResponseLogger defines what error goes to the log and what to display as JSON in client
func (e *ErrorBuilder) errResponseLogger() {
	// log.Errorf("Original error: %s\n", e.err)
	jsonParseErr := json.NewEncoder(e.w).Encode(e)
	if jsonParseErr != nil {
		log.Errorf("Json parse error: %s\n", jsonParseErr)
	}
}
