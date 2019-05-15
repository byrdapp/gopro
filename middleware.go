package main

import (
	"net/http"

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
		// ? If the JWT failed
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
				http.Error(w, http.ErrMissingContentLength.Error(), 400)
			}
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// * If JWT succeeded

		if !token.Valid {
			http.Error(w, "Token is not valid", http.StatusUnauthorized)
			return
		}
		// log.Info("Valid JWT")
		next(w, r)
	})
}
