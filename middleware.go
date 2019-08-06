package main

import (
	stdliberr "errors"
	"fmt"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/blixenkrone/gopro/utils/errors"

	jwt "github.com/dgrijalva/jwt-go"
)

// Claims -
type Claims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Claims   jwt.StandardClaims
}

const (
	// Time current token must be below until a refresh happens
	tokenRefreshThrottle = 10 * time.Minute
	// How much time will the token be extended for
	tokenExpirationTime = 30 * time.Minute
	proToken            = "pro_token"
)

// isJWTAuth middleware requires routes to possess a JWToken
func isJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := Claims{}
		c, err := r.Cookie("pro_token")
		if err != nil {
			if err == http.ErrNoCookie {
				fmt.Printf("Error %s", err.Error())
				errors.NewResErr(err, http.ErrNoCookie.Error(), http.StatusUnauthorized, w)
			}
			errors.NewResErr(err, "Error getting token", 503, w)
		}

		token, err := jwt.ParseWithClaims(c.Value, &claims.Claims, func(token *jwt.Token) (interface{}, error) {
			fmt.Println(c.Value)
			if sm, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				err := stdliberr.New("Invalid signing algorithm")
				errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
				spew.Dump(sm)
				return nil, err
			}
			return JWTSecret, nil
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
	tokenRefreshThrottle := time.Now().Add(tokenRefreshThrottle).Unix()
	fmt.Println(c.Claims.ExpiresAt)
	fmt.Println(tokenRefreshThrottle)
	if c.Claims.ExpiresAt < tokenRefreshThrottle {
		expirationTime := time.Now().Add(tokenExpirationTime)
		c.Claims.ExpiresAt = expirationTime.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, c.Claims)
		signedToken, err := token.SignedString(JWTSecret)
		if err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "pro_token",
			Value:  signedToken,
			Path:   "/",
			Secure: false,
		})
		fmt.Println("Refreshed token")
		return nil
	}
	return nil
}
