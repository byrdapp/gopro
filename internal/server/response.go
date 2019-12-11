package server

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type ResponseBuilder struct {
	Data   map[string]interface{} `json:"data"`
	Errors map[string]interface{} `json:"data"`
}

// ResponseBuilder builds custom errors to a http response writer
type ErrResponseBuilder struct {
	Status    int    `json:"status"`
	ClientMsg string `json:"msg"`
	w         http.ResponseWriter
	err       error
	traced    string
}

// NewResErr constructs and executes an err struct.
// Set stackTraced = "trace" to show error stack.
func NewResErr(err error, msg string, statusCode int, w http.ResponseWriter, stackTraced ...string) *ErrResponseBuilder {
	build := &ErrResponseBuilder{
		Status:    statusCode,
		ClientMsg: msg,
		w:         w,
		err:       err,
	}
	if len(stackTraced) > 0 {
		build.traced = stackTraced[0]
	}
	if build.traced == "trace" {
		build.errStackTraced()
	}
	if build.traced == "err" {
		log.Error(err)
	}
	if err := build.ErrorResponse(); err != nil {
		log.Errorf("Internal error: %s", err)
	}
	return build
}

type errResponse struct {
	Status    int    `json:"status"`
	ClientMsg string `json:"msg"`
}

func (r *ErrResponseBuilder) ErrorResponse() error {
	r.w.Header().Set("Content-Type", "application/json")
	errRes := &errResponse{
		Status:    r.Status,
		ClientMsg: r.ClientMsg,
	}
	r.w.WriteHeader(r.Status)
	return json.NewEncoder(r.w).Encode(errRes)
}

func (r *ErrResponseBuilder) errStackTraced() {
	// The `errors.Cause` function returns the originally wrapped error, which we can then type assert to its original struct type
	err := errors.WithStack(r.err)
	log.Errorf("originated: %+v", err)
}

// Imbed errors in the response JSON
func (r *ErrResponseBuilder) ErrorImbedded(err error, msg string, code int) *ErrResponseBuilder {
	return &ErrResponseBuilder{
		Status:    code,
		ClientMsg: msg,
		err:       err,
	}
}
