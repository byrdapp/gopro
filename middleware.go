package main

import (
	"fmt"
	"net/http"
	"os"
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
	proToken            = "pro_token"
	userToken           = "user_token"
)

func isAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		next(w, r)
	}
}

/**
 * Deprecated
 */
var isJWTAuth = func(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerToken := r.Header.Get(userToken)
		// ? verify here, that the user is a pro user

		if headerToken == "" {
			var err error
			err = fmt.Errorf("Headertoken value must not be empty or: '%s'", headerToken)
			errors.NewResErr(err, "No token or wrong token value provided", http.StatusUnauthorized, w)
			return
		}
		log.Infof("token type: %T", headerToken)
		_, err := fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			err = fmt.Errorf("Err: %s. Token: %s", err, headerToken)
			errors.NewResErr(err, "Error verifying token or token has expired", http.StatusFound, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		if os.Getenv("ENV") == "development" {
			log.Infoln("Middleware ran successfully")
		}
		next(w, r)
	})
}

func (c *Claims) refreshToken(w http.ResponseWriter) error {
	// tokenRefreshThrottle := time.Now().Add(tokenRefreshThrottle).Unix()
	// fmt.Println(c.Claims.ExpiresAt)
	// fmt.Println(tokenRefreshThrottle)
	// if c.Claims.ExpiresAt < tokenRefreshThrottle {
	// 	expirationTime := time.Now().Add(tokenExpirationTime)
	// 	c.Claims.ExpiresAt = expirationTime.Unix()
	// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c.Claims)
	// 	signedToken, err := token.SignedString(JWTSecretMust())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	http.SetCookie(w, &http.Cookie{
	// 		Name:   "pro_token",
	// 		Value:  signedToken,
	// 		Path:   "/",
	// 		Secure: false,
	// 	})
	// 	fmt.Println("Refreshed token")
	// 	return nil
	// }
	return nil
}
