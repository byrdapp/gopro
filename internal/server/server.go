package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mux "github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"

	storage "github.com/blixenkrone/gopro/internal/storage"
	firebase "github.com/blixenkrone/gopro/internal/storage/firebase"
	"github.com/blixenkrone/gopro/internal/storage/postgres"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var (
	log = logger.NewLogger()
	pq  storage.PQService
	fb  storage.FBService
)

// Server is used in main.go
type Server struct {
	HttpListenServer   *http.Server
	HttpRedirectServer *http.Server
	CertM              *autocert.Manager
	// handlermux http.Handler
}

// Creates a new server with HTTP2 & HTTPS
func NewServer() *Server {
	mux := mux.NewRouter()

	// mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./dist/pro-app/"))))
	// mux = mux.PathPrefix("/api/").Subrouter()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
		log.Infoln("Ran test")
		fmt.Fprintln(w, "Nothing to see here :-)")
	}).Methods("GET")
	mux.HandleFunc("/login", loginGetUserAccess).Methods("POST")

	// * Private endpoints
	mux.HandleFunc("/reauthenticate", isAuth(loginGetUserAccess)).Methods("GET")
	mux.HandleFunc("/secure", isAuth(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"msg": "Secure msg from gopro service"}`)); err != nil {
			log.Errorln(err)
		}
	})).Methods("GET")

	mux.HandleFunc("/admin/secure", isAdmin(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"msg": "Secure msg from gopro service to ADMINS!"}`)); err != nil {
			log.Errorln(err)
		}
	})).Methods("GET")

	mux.HandleFunc("/logoff", signOut).Methods("POST")

	mux.HandleFunc("/mail/send", isAuth(sendMail)).Methods("POST")
	mux.HandleFunc("/exif/image", isAuth(exifImages)).Methods("POST")
	mux.HandleFunc("/exif/video", isAuth(exifVideo)).Methods("POST")

	mux.HandleFunc("/profiles", isAuth(getProfiles)).Methods("GET")
	mux.HandleFunc("/profile/{id}", isAuth(getProfileByID)).Methods("GET")

	mux.HandleFunc("/auth/profile/token", isAuth(decodeTokenGetProfile)).Methods("GET")
	mux.HandleFunc("/profile/{id}", isAuth(getProProfile)).Methods("GET")

	mux.HandleFunc("/booking/upload", isAuth(bookingUploadToStorage)).Methods("POST")
	mux.HandleFunc("/booking/task/{uid}", isAuth(getBookingsByUID)).Methods("GET")
	mux.HandleFunc("/booking/task/{proUID}", isAuth(createBooking)).Methods("POST")
	mux.HandleFunc("/booking/task/{bookingID}", isAuth(updateBooking)).Methods("PUT")
	mux.HandleFunc("/booking/task/{bookingID}", isAuth(deleteBooking)).Methods("DELETE")
	mux.HandleFunc("/booking/task" /** isAdmin() middleware? */, isAuth(getProfileWithBookings)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200", "http://localhost:4201", "http://localhost", "https://pro.development.byrd.news", "https://pro.dev.byrd.news", "https://pro.byrd.news"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Accept", "Content-Length", "X-Requested-By", "user_token", "preview"},
		// AllowCredentials: true,
	})

	// https://medium.com/weareservian/automagical-https-with-docker-and-go-4953fdaf83d2
	// m := autocert.Manager{
	// 	Prompt:     autocert.AcceptTOS,
	// 	HostPolicy: autocert.HostWhitelist(host),
	// 	Cache:      autocert.DirCache("certs"),
	// }

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
	// httpSrv := &http.Server{
	// 	Addr:           ":http",
	// 	ReadTimeout:    httpsSrv.ReadTimeout,
	// 	WriteTimeout:   httpsSrv.WriteTimeout,
	// 	IdleTimeout:    httpsSrv.IdleTimeout,
	// 	MaxHeaderBytes: 1 << 20,
	// 	Handler:        mux,
	// 	// Handler:        m.HTTPHandler(nil),
	// }

	return &Server{
		HttpListenServer: httpsSrv,
		// HttpRedirectServer: httpSrv,
		// CertM:              &m,
	}
}

func (s *Server) InitDB() error {
	pqsrv, err := postgres.NewPQ()
	if err != nil {
		log.Fatalf("POSTGRESQL err: %s", err)
		return err
	}
	pq = pqsrv

	fbsrv, err := firebase.NewFB()
	if err != nil {
		log.Fatalf("Error starting firebase: %s", err)
		return err
	}
	fb = fbsrv
	return nil
}

func (s *Server) UseHTTP2() error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(s.HttpListenServer, &http2Srv)
	if err != nil {
		return err
	}
	log.Infoln("Using HTTP/2.0")
	return nil
}

func (s *Server) WaitForShutdown() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer pq.Close()
	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	log.Fatal(s.HttpListenServer.Shutdown(ctx))
	log.Println("Shutting down")
	os.Exit(0)
}
