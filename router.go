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
	mux.HandleFunc("/login", loginGetToken).Methods("POST")

	// * Private endpoints
	mux.HandleFunc("/reauthenticate", isAuth(loginGetToken)).Methods("GET")
	mux.HandleFunc("/secure", isAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"msg": "Secure msg from gopro service"}`))
	})).Methods("GET")

	mux.HandleFunc("/admin/secure", isAdmin(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"msg": "Secure msg from gopro service to ADMINS!"}`))
	})).Methods("GET")

	mux.HandleFunc("/logoff", signOut).Methods("POST")

	mux.HandleFunc("/mail/send", isAuth(mailtips.MailHandler)).Methods("POST")
	mux.HandleFunc("/slack/tip", isAuth(slack.PostSlackMsg)).Methods("POST")
	mux.HandleFunc("/exif", isAuth(getExif)).Methods("POST")
	mux.HandleFunc("/profiles", isAuth(getProfiles)).Methods("GET")
	mux.HandleFunc("/profile/{id}", isAuth(getProfileByID)).Methods("GET")

	mux.HandleFunc("/auth/profile/token", isAuth(decodeTokenGetProfile)).Methods("GET")
	mux.HandleFunc("/profile/{id}", isAuth(getProProfile)).Methods("GET")

	mux.HandleFunc("/booking/{uid}", isAuth(getBookingsByUID)).Methods("GET")
	mux.HandleFunc("/booking/{proUID}", isAuth(createBooking)).Methods("POST")
	mux.HandleFunc("/booking/{bookingID}", isAuth(updateBooking)).Methods("PUT")
	mux.HandleFunc("/booking/{bookingID}", isAuth(deleteBooking)).Methods("DELETE")
	mux.HandleFunc("/bookings" /** isAdmin() middleware? */, isAuth(getProfileWithBookings)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200", "http://localhost:4201"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "Content-Length", "X-Requested-By", "Set-Cookie", "user_token"},
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