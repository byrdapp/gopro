package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// ResponseBuilder builds custom errors to a http response writer
type ErrResponseBuilder struct {
	Code      int    `json:"code"`
	ClientMsg string `json:"msg"`
	w         http.ResponseWriter
	err       error
	traced    string
}

// NewResErr constructs and executes an err struct.
// Set stackTraced = "trace" to show error stack.
func NewResErr(err error, msg string, code int, w http.ResponseWriter, stackTraced ...string) *ErrResponseBuilder {
	build := &ErrResponseBuilder{
		Code:      code,
		ClientMsg: msg,
		w:         w,
		err:       err,
	}
	if len(stackTraced) > 0 {
		build.traced = stackTraced[0]
	}
	if build.traced == "trace" {
		stackErr := build.errStackTraced()
		log.Errorf("Error stacktraced: %as", stackErr)
	}
	if err := build.ErrorResponse(); err != nil {
		log.Errorf("Internal error: %s", err)
	}
	return build
}

// ErrorImbedded returns a responsebuilder interface, that will act as a errorresponse, if other responses are present i.e:
// {
// err: {msg:.., code: xxxx}, <- this
// data: {data},
// }
func (r *ErrResponseBuilder) ErrorImbedded(err error, msg string, code int) *ErrResponseBuilder {
	return &ErrResponseBuilder{
		Code:      code,
		ClientMsg: msg,
		err:       err,
	}
}

func (r *ErrResponseBuilder) ErrorResponse() error {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Code)
	return json.NewEncoder(r.w).Encode(r.ClientMsg)
}

func (r *ErrResponseBuilder) errStackTraced() error {
	err := errors.WithStack(r.err)
	return fmt.Errorf("Cause: %s", errors.Cause(err))
}
