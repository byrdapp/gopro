package server

import (
	"encoding/json"
	"net/http"

	"errors"
)

type simpleResponse struct {
	Msg  string `json:"msg,omitempty"`
	Code int    `json:"code,omitempty"`
}

type HttpStatusCode int

var ErrPanicRecover = errors.New("")
var ErrJSONEncoding = errors.New("json marshall encoding to byte array")
var ErrJSONDecoding = errors.New("json unmarshall decoding")
var ErrBadTokenHeader = errors.New("no or wrong token found in header")

const (
	_ = iota + 519
	StatusPanic
	// Marshall
	StatusJSONEncode
	// Unmarshall
	StatusJSONDecode
	StatusBadTokenHeader
)

var StatusText = map[HttpStatusCode]string{
	StatusJSONEncode:     ErrJSONEncoding.Error(),
	StatusJSONDecode:     ErrJSONDecoding.Error(),
	StatusBadTokenHeader: ErrBadTokenHeader.Error(),
}

func WriteClient(w http.ResponseWriter, code HttpStatusCode) (jsonerr HttpStatusCode) {
	enc := json.NewEncoder(w)
	res := &simpleResponse{
		Code: int(code),
		Msg:  code.StatusText(),
	}
	w.WriteHeader(int(code))
	if err := enc.Encode(res); err != nil {
		return StatusJSONEncode
	}
	return 0
}

func (code HttpStatusCode) StatusText() string {
	if http.StatusText(int(code)) != "" {
		return http.StatusText(int(code))
	} else {
		if val, ok := StatusText[code]; ok {
			return val
		} else {
			return "unknown error occurred internally - contact Simon on Slack."
		}
	}
}

func (code HttpStatusCode) LogError(err error) {
	log.Error(err)
	log.Errorf("ERR: %s ERR+V: %+v \n ClientMsg: %s", err, code.StatusText())
}

// ! Not in use
func (r *simpleResponse) errStackTraced(err error) {
	// The `errors.Cause` function returns the originally wrapped error, which we can then type assert to its original struct type
	nErr := errors.Unwrap(err)
	log.Errorf("OG err:", err)
	log.Errorf("unwrapped: %+v", nErr)
}
