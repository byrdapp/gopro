package server

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mux "github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/http2"

	"github.com/blixenkrone/gopro/internal/storage"
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
	HTTPListenServer *http.Server
	router           *mux.Router
	// HTTPRedirectServer *http.Server
	// certM  *autocert.Manager
	// handlermux http.Handler
}

// NewServer - Creates a new server with HTTP2 & HTTPS
func NewServer() *Server {
	r := mux.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:4200", "http://localhost:4201", "http://localhost", "https://pro.development.byrd.news", "https://pro.dev.byrd.news", "https://pro.byrd.news"},
		AllowedMethods: []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Accept", "Content-Length", "X-Requested-By", "user_token"},
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
		Handler: c.Handler(r),
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
		HTTPListenServer: httpsSrv,
		router:           r,
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

	defer pq.Close()
	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	log.Fatal(s.HTTPListenServer.Shutdown(ctx))
	log.Println("Shutting down")
	os.Exit(0)
}
