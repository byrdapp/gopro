package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
}

var jwtKey = []byte("thiskeyiswhat")

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

// Creates a new server with H2 & HTTPS
func newServer() (*Server, error) {
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
	mux.HandleFunc("/medias", isJWTAuth(getMedias))
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

	// Create server for redirecting HTTP to HTTPS
	httpSrv := &http.Server{
		Addr:         ":http",
		ReadTimeout:  httpsSrv.ReadTimeout,
		WriteTimeout: httpsSrv.WriteTimeout,
		IdleTimeout:  httpsSrv.IdleTimeout,
		Handler:      m.HTTPHandler(nil),
	}

	if err := useHTTP2(httpsSrv); err != nil {
		logrus.Warnf("Error with HTTP2 %s", err)
	}

	return &Server{
		httpsSrv: httpsSrv,
		httpSrv:  httpSrv,
		log:      newLogger(),
		certm:    &m,
	}, nil
}

func generateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var creds JWTCreds
		if err := decodeJSON(r.Body, creds); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		expirationTime := time.Now().Add(5 * time.Minute)
		claims := &Claims{
			Username: creds.Username,
			Claims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
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
			Expires: expirationTime,
			Value:   signedToken,
		})

		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			logrus.Warn(err)
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
