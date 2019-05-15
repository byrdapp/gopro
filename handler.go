package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/byblix/gopro/mailtips"
	"github.com/byblix/gopro/slack"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

var (
	host       = flag.String("host", "", "What host are you using?")
	production = flag.Bool("production", false, "Is it production?")
)

var jwtKey = []byte("thiskeyiswhat")

// Creates a new server with H2 & HTTPS
func newServer() error {
	mux := http.NewServeMux()
	// ? Public endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
		fmt.Fprintln(w, "Nothing to see here :-)")
	})
	mux.HandleFunc("/authenticate", generateJWT)

	// * Private endpoints
	mux.HandleFunc("/secure", isJWTAuth(secureMessage))
	mux.HandleFunc("/mail/send", isJWTAuth(mailtips.MailHandler))
	mux.HandleFunc("/slack/tip", isJWTAuth(slack.PostSlackMsg))
	mux.HandleFunc("/media", isJWTAuth(getMedias))
	mux.HandleFunc("/media/{id}", isJWTAuth(getMediaByID))
	mux.HandleFunc("/media", isJWTAuth(createMedia))

	// https://medium.com/weareservian/automagical-https-with-docker-and-go-4953fdaf83d2
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*host),
		Cache:      autocert.DirCache("/certs"),
	}

	httpsSrv := &http.Server{
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Addr:              ":https",
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
		Handler: mux,
	}

	// Serve on localhost with localhost certs if no host provided
	if *host == "" {
		httpsSrv.Addr = "localhost:8085"
		log.Info("Serving on http://localhost:8085")
		// log.Fatal(httpsSrv.ListenAndServeTLS("./certs/insecure_cert.pem", "./certs/insecure_key.pem"))
		if err := httpsSrv.ListenAndServe(); err != nil {
			return err
		}
	}

	// Create server for redirecting HTTP to HTTPS
	httpSrv := &http.Server{
		Addr:         ":http",
		ReadTimeout:  httpsSrv.ReadTimeout,
		WriteTimeout: httpsSrv.WriteTimeout,
		IdleTimeout:  httpsSrv.IdleTimeout,
		Handler:      m.HTTPHandler(nil),
	}

	if err := useHTTP2(httpsSrv); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	httpsSrv.TLSConfig.GetCertificate = m.GetCertificate
	log.Info("Serving on https, authenticating for https://", *host)
	if err := httpsSrv.ListenAndServeTLS("", ""); err != nil {
		return err
	}
	return nil
}

// isJWTAuth middleware requires routes to possess a JWToken
func isJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		// ? If the JWT failed
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, http.ErrNoCookie.Error(), http.StatusUnauthorized)
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// * If JWT succeeded
		token, err := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				http.Error(w, http.ErrMissingContentLength.Error(), 400)
			}
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if token.Valid {
			log.Info("Valid JWT")
			next(w, r)
		}
	})
}

// JWTClaims -
type JWTClaims struct {
	Username string `json:"username"`
	Claims   jwt.StandardClaims
}

func generateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var creds JWTCreds
		if err := decodeJSON(r.Body, creds); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		expirationTime := time.Now().Add(5 * time.Minute).Unix()
		claims := &JWTClaims{
			Username: creds.Username,
			Claims: jwt.StandardClaims{
				ExpiresAt: expirationTime,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.Claims)
		signedToken, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func secureMessage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Secret hello from go-pro!"))
}

func useHTTP2(httpsSrv *http.Server) error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(httpsSrv, &http2Srv)
	if err != nil {
		return err
	}
	return nil
}

// JWTCreds for at user to get JWT
type JWTCreds struct {
	Username string `json:"username"`
}

func decodeJSON(r io.Reader, val JWTCreds) error {
	err := json.NewDecoder(r).Decode(&val)
	if err != nil {
		return err
	}
	return nil
}
