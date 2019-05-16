package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	mux "github.com/gorilla/mux"

	"github.com/byblix/gopro/mailtips"
	"github.com/byblix/gopro/slack"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

type Server struct {
	httpsSrv *http.Server
	httpSrv  *http.Server
	certm    *autocert.Manager
	log      *logrus.Logger
	mux      *mux.Router
}

var jwtKey = []byte("thiskeyiswhat")

// Creates a new server with H2 & HTTPS
func newServer() *Server {
	mux := mux.NewRouter()
	// ? Public endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
		fmt.Fprintln(w, "Nothing to see here :-)")
	}).Methods("GET")
	mux.HandleFunc("/authenticate", generateJWT).Methods("POST")
	mux.HandleFunc("/reauthenticate", isJWTAuth(generateJWT)).Methods("GET")

	// * Private endpoints
	mux.HandleFunc("/secure", isJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Secure msg from gopro service"))
	})).Methods("GET")

	mux.HandleFunc("/mail/send", isJWTAuth(mailtips.MailHandler)).Methods("POST")
	mux.HandleFunc("/slack/tip", isJWTAuth(slack.PostSlackMsg)).Methods("POST")
	mux.HandleFunc("/media", isJWTAuth(getMedias)).Methods("GET")
	mux.HandleFunc("/media/{id}", isJWTAuth(getMediaByID)).Methods("GET")
	mux.HandleFunc("/media", isJWTAuth(createMedia)).Methods("POST")

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

	// Create server for redirecting HTTP to HTTPS
	httpSrv := &http.Server{
		Addr:         ":http",
		ReadTimeout:  httpsSrv.ReadTimeout,
		WriteTimeout: httpsSrv.WriteTimeout,
		IdleTimeout:  httpsSrv.IdleTimeout,
		Handler:      m.HTTPHandler(nil),
	}

	return &Server{
		httpsSrv: httpsSrv,
		httpSrv:  httpSrv,
		log:      newLogger(),
		certm:    &m,
	}
}

func generateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var creds JWTCreds
		if err := decodeJSON(r.Body, creds); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		tokenExpirationTime := time.Now().Add(time.Second * 20)
		claims := &Claims{
			Username: creds.Username,
			Claims: jwt.StandardClaims{
				ExpiresAt: tokenExpirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.Claims)
		signedToken, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Expires: time.Now().Add(time.Minute * 1),
			Value:   signedToken,
		})

		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			logrus.Warn(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func newLogger() *logrus.Logger {
	logger := logrus.StandardLogger()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:    true,
		FullTimestamp:  true,
		DisableSorting: true,
	})
	return logger
}

func (s *Server) useHTTP2() error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(s.httpsSrv, &http2Srv)
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
