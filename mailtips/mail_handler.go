package mailtips

import (
	"encoding/json"
	"net/http"
)

// MailHandler handles mail requests
// /v1/mail/send + body
func MailHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := SendMail(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(resp.StatusCode)
	b, _ := json.Marshal(resp.StatusCode)
	w.Write(b)
}
