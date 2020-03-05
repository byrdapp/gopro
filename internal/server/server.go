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

	mux "github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/http2"

	"github.com/blixenkrone/byrd/byrd-pro-api/internal/storage"
	firebase "github.com/blixenkrone/byrd/byrd-pro-api/internal/storage/firebase"
	"github.com/blixenkrone/byrd/byrd-pro-api/internal/storage/postgres"
	"github.com/blixenkrone/byrd/byrd-pro-api/pkg/logger"
)

var (
	log = logger.NewLogger()
	pq  *postgres.Queries
	fb  storage.FBService
)

// ! TODO: Same approach with FB db as postgres

// Server is used in main.go
type Server struct {
	HTTPListenServer *http.Server
	router           *mux.Router
}

// NewServer - Creates a new server with HTTP2 & HTTPS
func NewServer() *Server {
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
		Addr:              ":https",
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
		Handler: c.Handler(r),
	}

	return &Server{
		HTTPListenServer: httpsSrv,
		router:           r,
	}
}

func (s *Server) InitDB() error {
	conn, err := sql.Open("postgres", os.Getenv("POSTGRES_CONNSTR"))
	if err != nil {
		return err
	}
	db := postgres.New(conn)
	pq = db

	fbsrv, err := firebase.NewFB()
	if err != nil {
		return err
	}
	fb = fbsrv
	return nil
}

func (s *Server) UseHTTP2() error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(s.HTTPListenServer, &http2Srv)
	if err != nil {
		return err
	}
	log.Infoln("Using HTTP/2.0")
	return nil
}

func (s *Server) WaitForShutdown() {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		if err := pq.Close(); err != nil {
			log.Error(err)
		}
	}()
	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	log.Fatal(s.HTTPListenServer.Shutdown(ctx))
	log.Println("Shutting down")
	os.Exit(0)
}
