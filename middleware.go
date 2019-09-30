package main

import (
	"fmt"
	"net/http"

	resErr "github.com/blixenkrone/gopro/utils/errors"
)

const (
	userToken    = "user_token"
	isAdminClaim = "is_admin"
)

var isAdmin = func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerToken := r.Header.Get(userToken)
		if headerToken == "" {
			err := fmt.Errorf("Headertoken value must not be empty or: '%s'", headerToken)
			resErr.NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}

		token, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			resErr.NewResErr(err, "Token could not be verified, or the token is expired.", http.StatusUnauthorized, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		if ok, err := fb.IsAdminUID(r.Context(), token.UID); ok && err == nil {
			next(w, r)
			return
		}
		if err != nil {
			http.Error(w, "Error getting admin rights", http.StatusBadRequest)
		}
		err = fmt.Errorf("No admin rights found")
		resErr.NewResErr(err, err.Error(), http.StatusBadRequest, w)
	}
}

var isAuth = func(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerToken := r.Header.Get(userToken)
		// ? verify here, that the user is a pro user

		if headerToken == "" {
			err := fmt.Errorf("Headertoken value must not be empty or: '%s'", headerToken)
			resErr.NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}
		tkn, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			err = fmt.Errorf("Err: %s", err)
			resErr.NewResErr(err, "Error verifying token or token has expired", http.StatusUnauthorized, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		log.Info(tkn.UID)
		next(w, r)
	})
}
