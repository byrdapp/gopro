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
	msg, ok := code.StatusText()
	res := &simpleResponse{
		Code: int(code),
		Msg:  msg,
	}
	if !ok {
		w.WriteHeader(int(res.Code))
	}
	if err := enc.Encode(res); err != nil {
		return StatusJSONEncode
	}
	return 0
}

func (code HttpStatusCode) StatusText() (string, bool) {
	if http.StatusText(int(code)) != "" {
		return http.StatusText(int(code)), true
	} else {
		if val, ok := StatusText[code]; ok {
			return val, ok
		} else {
			return "unknown error occurred internally - contact Simon on Slack.", false
		}
	}
}

func (code HttpStatusCode) LogError(err error) {
	log.Error(err)
	msg, ok := code.StatusText()
	log.Errorf("err val: %+v \n ClientMsg: %s (statuscode anticipated and defined in API?: %v)", err, msg, ok)
}
