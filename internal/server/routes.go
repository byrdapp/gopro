package server

import (
	"fmt"
	"net/http"
)

func (s *Server) InitRoutes() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooEarly)
		log.Infoln("Ran test")
		fmt.Fprintln(w, "Nothing to see here :-)")
	}).Methods("GET")
	s.router.HandleFunc("/login", loginGetUserAccess).Methods("POST")

	// * Private endpoints
	s.router.HandleFunc("/reauthenticate", isAuth(loginGetUserAccess)).Methods("GET")
	s.router.HandleFunc("/secure", isAuth(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"msg": "Secure msg from byrd-pro-api service"}`)); err != nil {
			log.Errorln(err)
		}
	})).Methods("GET")

	s.router.HandleFunc("/admin/secure", isAdmin(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`{"msg": "Secure msg from byrd-pro-api service to ADMINS!"}`)); err != nil {
			log.Errorln(err)
		}
	})).Methods("GET")

	s.router.HandleFunc("/logoff", signOut).Methods("POST")

	s.router.HandleFunc("/mail/send", recoverFunc(isAuth(sendMail))).Methods("POST")
	s.router.HandleFunc("/exif/image", recoverFunc(isAuth(exifImages))).Methods("POST")
	s.router.HandleFunc("/exif/video", recoverFunc(isAuth(exifVideo))).Methods("POST")

	s.router.HandleFunc("/profiles", isAuth(getProfiles)).Methods("GET")
	s.router.HandleFunc("/profile/{id}", isAuth(getProfileByID)).Methods("GET")

	s.router.HandleFunc("/auth/profile/token", isAuth(decodeTokenGetProfile)).Methods("GET")
	s.router.HandleFunc("/profile/{id}", isAuth(getProProfile)).Methods("GET")

	s.router.HandleFunc("/booking/task/{uid}", isAuth(getBookingsByUID)).Methods("GET")
	s.router.HandleFunc("/booking/task/{proUID}", isAuth(createBooking)).Methods("POST")
	s.router.HandleFunc("/booking/task/{bookingID}", isAuth(updateBooking)).Methods("PUT")
	s.router.HandleFunc("/booking/task/{bookingID}", isAuth(deleteBooking)).Methods("DELETE")
	// s.router.HandleFunc("/booking/task" /** isAdmin() middleware? */, isAuth(getProfileWithBookings)).Methods("GET")
}
