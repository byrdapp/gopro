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

// isJWTAuth middleware requires routes to possess a JWToken
func isJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		fmt.Printf("Token: %s\n", c.Value)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, http.ErrNoCookie.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		claims := Claims{}
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !token.Valid {
			http.Error(w, "Token is not valid", http.StatusUnauthorized)
			return
		}
		// TODO Make the token refresh
		if err := claims.refreshToken(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next(w, r)
	})
}

func (c *Claims) isRefreshToken() {
	// c.Claims.ExpiresAt
}

func (c *Claims) refreshToken(w http.ResponseWriter) error {
	tokenRefreshThrottle := time.Now().Add(10 * time.Second).Unix()
	if tokenRefreshThrottle > c.Claims.ExpiresAt {
		expirationTime := time.Now().Add(10 * time.Second)
		c.Claims.ExpiresAt = expirationTime.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, c.Claims)
		signedToken, err := token.SignedString(jwtKey)
		if err != nil {
			return err
		}
		// TODO Token is not refreshed
		http.SetCookie(w, &http.Cookie{
			Name:  "token",
			Value: signedToken,
		})
		fmt.Println("Refreshed token")
		return nil
	}
	return nil
}
