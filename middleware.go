package main

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// JWTClaims -
type Claims struct {
	Username string `json:"username"`
	Claims   jwt.StandardClaims
}

const (
	// Time current token must be below until a refresh happens
	TOKEN_REFRESH_TROTTLE = 10 * time.Second
	// How much time will the token be extended for
	TOKEN_EXPIRATION_TIME = 30 * time.Second
)

// isJWTAuth middleware requires routes to possess a JWToken
func isJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		claims := Claims{}
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, http.ErrNoCookie.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("Token: %s\n", c.Value)

		token, err := jwt.ParseWithClaims(c.Value, &claims.Claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				http.Error(w, "Invalid signing algorithm", 401)
			}
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		// * refresh if the token is expired but value still in cookie
		if err := claims.refreshToken(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !token.Valid {
			http.Error(w, "Token is not valid", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}

func (c *Claims) refreshToken(w http.ResponseWriter) error {
	tokenRefreshThrottle := time.Now().Add(TOKEN_REFRESH_TROTTLE).Unix()
	fmt.Println(c.Claims.ExpiresAt)
	fmt.Println(tokenRefreshThrottle)
	if c.Claims.ExpiresAt < tokenRefreshThrottle {
		expirationTime := time.Now().Add(TOKEN_EXPIRATION_TIME)
		c.Claims.ExpiresAt = expirationTime.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, c.Claims)
		signedToken, err := token.SignedString(jwtKey)
		if err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "token",
			Value: signedToken,
		})
		fmt.Println("Refreshed token")
		return nil
	}
	return nil
}
