package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/byrdapp/byrd-pro-api/internal/slack"
)

const (
	userToken = "user_token"
)

func (s *server) loggerMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("<< %s %s %v\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *server) recoverFunc(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.writeClient(w, http.StatusInternalServerError).LogError(err.(error))

				var recoverReason string
				switch recovered := err.(type) {
				case error:
					recoverReason = recovered.Error()
				case string:
					recoverReason = recovered
				default:
					recoverReason = recovered.(error).Error()
				}

				if os.Getenv("PANIC_NOTIFICATIONS") == "true" {
					prf, err := s.fb.GetProfileByToken(r.Context(), r.Header.Get("user_token"))
					if err != nil {
						s.Errorf("profile was not found / header not present")
					}
					msg := fmt.Sprintf("%s (%s) messed up route: %s. reason might be: %v",
						prf.DisplayName, prf.UserID, r.URL.String(), recoverReason)

					slack.Hook(msg, prf.UserPicture).Panic()
				}
				return
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (s *server) isAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		headerToken := r.Header.Get(userToken)
		if headerToken == "" {
			s.writeClient(w, StatusBadTokenHeader)
			return
		}

		token, err := s.fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			s.writeClient(w, StatusBadTokenHeader)
			return
		}

		if ok, err := s.fb.IsAdminUID(r.Context(), token.UID); ok && err == nil {
			next(w, r)
			return
		}
		err = errors.New("No admin rights found:")
		s.writeClient(w, http.StatusBadRequest).LogError(err)
	}
}

func (s *server) isAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		headerToken := r.Header.Get(userToken)
		// ? verify here, that the user is a pro user
		if headerToken == "" {
			s.writeClient(w, StatusBadTokenHeader)
			return
		}
		token, err := s.fb.VerifyToken(r.Context(), headerToken)
		if err != nil {
			s.writeClient(w, StatusBadTokenHeader)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}

		isPro, err := s.fb.IsProfessional(r.Context(), token.UID)
		if err != nil {
			http.RedirectHandler("/login", http.StatusBadRequest)
			return
		}
		if !isPro {
			s.writeClient(w, http.StatusUnauthorized)
			http.RedirectHandler("/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
