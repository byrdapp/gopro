package main

import (
	"fmt"
	"net/http"
	"time"

	"firebase.google.com/go/auth"

	"github.com/blixenkrone/gopro/utils/errors"

	jwt "github.com/dgrijalva/jwt-go"
)

// Claims -
type Claims struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	FbClaims  auth.Token
	JWTClaims jwt.StandardClaims
	UID       string
}

const (
	// Time current token must be below until a refresh happens
	tokenRefreshThrottle = 10 * time.Minute
	// How much time will the token be extended for
	tokenExpirationTime = 30 * time.Minute
	userToken           = "user_token"
	isAdminClaim        = "is_admin"
)

var isAdmin = func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerToken := r.Header.Get(userToken)
		if headerToken == "" {
			err := fmt.Errorf("Headertoken value must not be empty or: '%s'", headerToken)
			errors.NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}

		token, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			errors.NewResErr(err, "Token could not be verified, or the token is expired.", http.StatusUnauthorized, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		if ok := fb.IsAdminClaims(token.Claims); ok {
			next(w, r)
			return
		}
		err = fmt.Errorf("No admin rights found")
		errors.NewResErr(err, err.Error(), http.StatusBadRequest, w, "trace")
		return
	}
}

var isAuth = func(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerToken := r.Header.Get(userToken)
		// ? verify here, that the user is a pro user

		if headerToken == "" {
			err := fmt.Errorf("Headertoken value must not be empty or: '%s'", headerToken)
			errors.NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}
		_, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			err = fmt.Errorf("Err: %s", err)
			errors.NewResErr(err, "Error verifying token or token has expired", http.StatusUnauthorized, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		next(w, r)
	})
}
