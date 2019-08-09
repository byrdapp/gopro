package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/blixenkrone/gopro/mailtips"
	"github.com/blixenkrone/gopro/slack"
	mux "github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
)

// Server is used in main.go
type Server struct {
	httpsSrv *http.Server
	httpSrv  *http.Server
	certm    *autocert.Manager
	// handlermux http.Handler
}

// Creates a new server with HTTP2 & HTTPS
func newServer() *Server {
	mux := mux.NewRouter()
	// ? Public endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
		fmt.Fprintln(w, "Nothing to see here :-)")
	}).Methods("GET")
	mux.HandleFunc("/authenticate", loginGetToken).Methods("POST")
	mux.HandleFunc("/securion/plans", getSecurionPlans).Methods("GET")
	mux.HandleFunc("/login", loginGetToken).Methods("POST")

	// * Private endpoints
	mux.HandleFunc("/reauthenticate", isJWTAuth(loginGetToken)).Methods("GET")
	mux.HandleFunc("/secure", isJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Secure msg from gopro service"))
	})).Methods("GET")

	mux.HandleFunc("/logoff", signOut).Methods("POST")

	mux.HandleFunc("/mail/send", isJWTAuth(mailtips.MailHandler)).Methods("POST")
	mux.HandleFunc("/slack/tip", isJWTAuth(slack.PostSlackMsg)).Methods("POST")
	mux.HandleFunc("/exif", isJWTAuth(getExif)).Methods("POST")
	mux.HandleFunc("/medias", isJWTAuth(getMedias)).Methods("GET")
	mux.HandleFunc("/media/{id}", isJWTAuth(getMediaByID)).Methods("GET")
	mux.HandleFunc("/media", isJWTAuth(createMedia)).Methods("POST")

	mux.HandleFunc("/profile/decode", isJWTAuth(decodeTokenGetProfile)).Methods("GET")
	mux.HandleFunc("/profile/{id}", isJWTAuth(getProProfile)).Methods("GET")
	mux.HandleFunc("/stats/{id}", isJWTAuth(getProStats)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "Content-Length", "X-Requested-By", "Set-Cookie", "user_token", "pro_token"},
		AllowCredentials: true,
	})

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
		MaxHeaderBytes:    1 << 20,
		Addr:              ":https",
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
		Handler: c.Handler(mux),
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
		certm:    &m,
	}
}

func (s *Server) useHTTP2() error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(s.httpsSrv, &http2Srv)
	if err != nil {
		return err
	}
	return nil
}
