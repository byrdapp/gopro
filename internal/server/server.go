package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	mux "github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/http2"

	"github.com/byrdapp/byrd-pro-api/internal/storage"
	firebase "github.com/byrdapp/byrd-pro-api/internal/storage/firebase"
	"github.com/byrdapp/byrd-pro-api/internal/storage/postgres"
)

var (
// log = logger.NewLogger()
// pq  *postgres.Queries
// fb storage.FBService
)

// type server struct {
// 	srv    *http.Server
// 	router *mux.Router
// 	db     *database
// 	logger
// }

type logger interface {
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

// ! TODO: Same approach with FB db as postgres

// server is used in main.go
type server struct {
	srv    *http.Server
	router *mux.Router
	pq     *postgres.Queries
	fb     storage.FBService
	logger
}

// NewServer - Creates a new server with HTTP2 & HTTPS
func NewServer() (*server, error) {
	r := mux.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200", "http://localhost:4201", "http://localhost", "https://pro.development.byrd.news", "https://pro.dev.byrd.news", "https://pro.byrd.news"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Accept", "Content-Length", "X-Requested-By", "user_token"},
	})

	httpsSrv := &http.Server{
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		Addr:              ":3000",
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
		Handler: c.Handler(r),
	}

	conn, err := sql.Open("postgres", os.Getenv("POSTGRES_CONNSTR"))
	if err != nil {
		return nil, err
	}
	pq := postgres.New(conn)

	fbsrv, err := firebase.NewFB()
	if err != nil {
		return nil, err
	}

	return &server{
		srv:    httpsSrv,
		router: r,
		pq:     pq,
		fb:     fbsrv,
		logger: logrus.New(),
	}, nil
}

func (s *server) Routes() {
	// s.router.Use(s.recoverFunc)
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
	}).Methods("GET")
	s.router.HandleFunc("/login", s.loginGetUserAccess()).Methods("POST")

	// * Private endpoints
	s.router.HandleFunc("/reauthenticate", s.isAuth(s.loginGetUserAccess())).Methods("GET")
	s.router.HandleFunc("/secure", s.isAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"msg": "Secure msg from byrd-pro-api service"}`))
	})).Methods("GET")

	s.router.HandleFunc("/admin/secure", s.isAdmin(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"msg": "Secure msg from byrd-pro-api service to ADMINS!"}`))
	})).Methods("GET")

	s.router.HandleFunc("/logoff", signOut).Methods("POST")
	s.router.HandleFunc("/mail/send", s.isAuth(s.sendMail())).Methods("POST")
	s.router.HandleFunc("/exif/image", s.isAuth(s.exifImages())).Methods("POST")
	s.router.HandleFunc("/exif/video", s.isAuth(s.exifVideo())).Methods("POST")

	s.router.HandleFunc("/profiles", s.isAuth(s.getProfiles())).Methods("GET")
	s.router.HandleFunc("/profile/{id}", s.isAuth(s.getProfileByID())).Methods("GET")

	s.router.HandleFunc("/auth/profile/token", s.isAuth(s.decodeTokenGetProfile())).Methods("GET")
	s.router.HandleFunc("/profile/{id}", s.isAuth(s.getProProfile())).Methods("GET")

	s.router.HandleFunc("/booking/task/{uid}", s.isAuth(s.getBookingsByUID())).Methods("GET")

	s.router.HandleFunc("/booking/task", s.isAuth(s.createBooking())).Methods("POST")
	s.router.HandleFunc("/booking/accepted", s.isAuth(s.acceptBooking())).Methods("PUT") // ==> update accepted true/false
	// s.router.HandleFunc("/booking/task/{proUID}", s.isAuth(createSpecficBooking)).Methods("POST")

	s.router.HandleFunc("/booking/task/{bookingID}", s.isAuth(s.updateBooking())).Methods("PUT")
	s.router.HandleFunc("/booking/task/{bookingID}", s.isAuth(s.deleteBooking())).Methods("DELETE")
	// s.router.HandleFunc("/booking/task" /** isAdmin() middleware? */, isAuth(getProfileWithBookings)).Methods("GET")
}

func (s *server) UseHTTP2() error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(s.srv, &http2Srv)
	if err != nil {
		return err
	}
	s.Infof("Using HTTP/2.0")
	return nil
}

func (s *server) WaitForShutdown() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		if err := s.pq.Close(); err != nil {
			s.Errorf("%v", err)
		}
	}()
	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	s.Fatalf("%v", s.srv.Shutdown(ctx))
	s.Infof("Shutting down")
	os.Exit(0)
}

func (s *server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}
