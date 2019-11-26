package server

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
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
			NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}

		token, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			NewResErr(err, "Token could not be verified, or the token is expired.", http.StatusUnauthorized, w)
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
		NewResErr(err, err.Error(), http.StatusBadRequest, w)
	}
}

var isAuth = func(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		headerToken := r.Header.Get(userToken)
		// ? verify here, that the user is a pro user
		if headerToken == "" {
			err := errors.Errorf("header token empty or wrong format: '%s'", headerToken)
			NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}
		token, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			err := errors.Cause(err)
			NewResErr(err, "Error verifying token or token has expired", http.StatusUnauthorized, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}

		if isPro, err := fb.IsProfessional(r.Context(), token.UID); !isPro || err != nil {
			NewResErr(err, err.Error(), http.StatusUnauthorized, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}

		next(w, r)
	})
}
