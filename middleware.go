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
	proToken            = "pro_token"
)

// isJWTAuth middleware requires routes to possess a JWToken
var isJWTAuthFB = func(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := r.Cookie(proToken)
		if err != nil {
			if err == http.ErrNoCookie {
				fmt.Printf("Error %s", err.Error())
				// force user to relogin
				errors.NewResErr(err, http.ErrNoCookie.Error(), http.StatusUnauthorized, w)
				http.RedirectHandler("/login", http.StatusFound)
			}
			errors.NewResErr(err, "Error getting token", 503, w)
			http.RedirectHandler("/login", http.StatusFound)
		}
		token, err := fb.VerifyToken(r.Context(), cookie.Value)
		if err != nil {
			errors.NewResErr(err, "Error verifying token", http.StatusFound, w)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		_ = token
		// ? refresh if the token is expired but value still in cookie
		next(w, r)
	})
}

func isAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		next(w, r)
	}
}

/**
 * Deprecated
 */
func isJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := Claims{}
		cookie, err := r.Cookie(proToken)
		if err != nil {
			if err == http.ErrNoCookie {
				fmt.Printf("Error %s", err.Error())
				errors.NewResErr(err, http.ErrNoCookie.Error(), http.StatusUnauthorized, w)
			}
			errors.NewResErr(err, "Error getting token", 503, w)
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &claims.JWTClaims, func(token *jwt.Token) (interface{}, error) {
			fmt.Println(cookie.Value)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
				return nil, err
			}
			return JWTSecretMust(), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				errors.NewResErr(err, "Invalid signature", 503, w)
				return
			}
			errors.NewResErr(err, "Error parsing Claims for JWT", http.StatusForbidden, w)
			return
		}
		// ? refresh if the token is expired but value still in cookie
		// if err := claims.refreshToken(w); err != nil {
		// 	errors.NewResErr(err, "Error refreshing token", 503, w)
		// 	return
		// }
		if !token.Valid {
			http.Error(w, "Token is not valid", http.StatusUnauthorized)
			return
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
