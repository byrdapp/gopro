package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"

	exif "github.com/blixenkrone/gopro/internal/exif"
	image "github.com/blixenkrone/gopro/internal/exif/image"
	video "github.com/blixenkrone/gopro/internal/exif/video"
	mail "github.com/blixenkrone/gopro/internal/mail"
	storage "github.com/blixenkrone/gopro/internal/storage"
	conversion "github.com/blixenkrone/gopro/pkg/conversion"
	timeutil "github.com/blixenkrone/gopro/pkg/time"
)

var JSONEncodingError = errors.New("Error converting exif to JSON")

var signOut = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "application/json")
		http.SetCookie(w, &http.Cookie{
			Name:   "user_token",
			Value:  "",
			MaxAge: 0,
		})
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

// Credentials for at user to get JWT
type Credentials struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

var loginGetToken = func(w http.ResponseWriter, r *http.Request) {
	// ? verify here, that the user is a pro user
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		var err error
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			NewResErr(err, "Error decoding JSON from request body", http.StatusBadRequest, w)
			return
		}
		defer r.Body.Close()
		if creds.Password == "" || creds.Email == "" {
			err := fmt.Errorf("Missing email or password in credentials")
			NewResErr(err, err.Error(), http.StatusInternalServerError, w)
			return
		}

		usr, err := fb.GetProfileByEmail(r.Context(), creds.Email)
		if err != nil {
			NewResErr(err, "Error finding authentication for profile. Is the email/password correct, and does the user exist?", http.StatusBadRequest, w)
			return
		}

		if isPro, err := fb.IsProfessional(r.Context(), usr.UID); !isPro || err != nil {
			NewResErr(err, err.Error(), http.StatusUnauthorized, w)
			return
		}

		// Is user an admin? Set claims as such.
		claims := make(map[string]interface{})
		isAdmin, err := fb.IsAdminUID(r.Context(), usr.UID)
		if err != nil {
			NewResErr(err, "Error admin ref was not found", http.StatusBadRequest, w)
			return
		}
		claims[isAdminClaim] = isAdmin
		signedToken, err := fb.CreateCustomTokenWithClaims(r.Context(), usr.UID, claims)
		if err != nil {
			NewResErr(err, "Error creating token!", http.StatusInternalServerError, w, "trace")
			return
		}

		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			NewResErr(err, "Error encoding JSON token", http.StatusInternalServerError, w)
			return
		}
	}
}

// /profile/decode func attempts to return a profile from a given client UID header
var decodeTokenGetProfile = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var err error
		clientToken := r.Header.Get(userToken)
		if clientToken == "" {
			err = fmt.Errorf("No header token from client")
			NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}
		fbtoken, err := fb.VerifyToken(r.Context(), clientToken)
		if err != nil {
			NewResErr(err, "No token provided in headers", http.StatusBadRequest, w)
			return
		}
		profile, err := fb.GetProfile(r.Context(), fbtoken.UID)
		if err != nil {
			NewResErr(err, "Error getting profile", http.StatusInternalServerError, w)
			return
		}

		if err := json.NewEncoder(w).Encode(profile); err != nil {
			NewResErr(err, "Error encoding JSON token", http.StatusInternalServerError, w)
			return
		}
	}
}

// /profile/{uid}
var getProfileByID = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		params := mux.Vars(r)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		val, err := fb.GetProfile(ctx, params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err := json.NewEncoder(w).Encode(val); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// getProfiles endpoint: /profiles
var getProfiles = func(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("content-type", "application/json")
	medias, err := fb.GetProfiles(r.Context())
	if err != nil {
		NewResErr(err, "Error finding media profiles", http.StatusFound, w)
	}
	if err := json.NewEncoder(w).Encode(medias); err != nil {
		NewResErr(err, "JSON Encoding failed", 500, w)
	}
}

// getExif recieves body with img files
// it attempts to fetch EXIF data from each image
// if no exif data, the error message will be added to the response without breaking out of the loop until EOF
var exifImages = func(w http.ResponseWriter, r *http.Request) {
	// r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "multipart/form-data")
		_, cancel := context.WithTimeout(r.Context(), time.Duration(time.Second*10))
		defer cancel()

		// Parse media type to get type of media
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			NewResErr(err, "Could not parse request body", http.StatusBadRequest, w)
			return
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(r.Body, params["boundary"])
			defer r.Body.Close()
			var res []*exif.Output
			for {
				// read length of files
				var x exif.Output
				part, err := mr.NextPart()
				if err != nil {
					if err == io.EOF {
						break
					}
					NewResErr(err, "error reading file: "+part.FileName(), http.StatusBadRequest, w)
					break
				}
				output := image.ReadImage(part)

				x = *output
				res = append(res, &x)
			}

			if err := json.NewEncoder(w).Encode(res); err != nil {
				NewResErr(err, JSONEncodingError.Error(), http.StatusInternalServerError, w)
				return
			}
		}
	}
}

var exifVideo = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			NewResErr(err, err.Error(), http.StatusNotFound, w, "err")
			return
		}
		w.Header().Set("Content-Type", mediaType)

		video, err := video.ReadVideo(r.Body)
		if err != nil {
			NewResErr(err, err.Error(), http.StatusNotFound, w, "err")
			return
		}

		out := video.CreateVideoExifOutput()
		defer func() {
			if err := video.File.Close(); err != nil {
				log.Errorln(err)
			}
			if err := video.File.RemoveFile(); err != nil {
				log.Error(err)
			}
			if err := r.Body.Close(); err != nil {
				log.Error(err)
			}
		}()

		if err := json.NewEncoder(w).Encode(out); err != nil {
			NewResErr(err, JSONEncodingError.Error(), http.StatusInternalServerError, w, "trace")
		}
	}
}

type ClientReq struct {
	Geoloc           map[string]int `json:"_geoloc"`
	IsHash           bool           `json:"isHash"`
	StoryHeadline    string         `json:"storyHeadline"`
	StoryDescription string         `json:"storyDescription"`
	Media            []*media       `json:"media"`
}

type media struct {
	Geoloc         map[string]int `json:"_geoloc"`
	MediaDate      int64          `json:"mediaDate"`
	MediaExtension string         `json:"mediaExtension"`
	MediaType      string         `json:"mediaType"`
	MediaDevice    string         `json:"mediaDevice"`
	IsVerified     bool           `json:"isVerified"`
	MediaWidth     int            `json:"mediaWidth"`
	MediaHeight    int            `json:"mediaHeight"`
	MediaSize      float64        `json:"mediaSize"`
}

var uploadByrdAPIRequest = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "multipart/form-data")
		_, cancel := context.WithTimeout(r.Context(), time.Duration(time.Second*10))
		defer cancel()

		// Parse media type to get type of media
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			NewResErr(err, "Could not parse request body", http.StatusBadRequest, w)
			return
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(r.Body, params["boundary"])
			defer r.Body.Close()
			for {
				part, err := mr.NextPart()
				// defer part.Close()
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Error(err)
					break
				}
				fileName := part.FormName()
				formVal := r.FormValue(fileName)
				log.Info(formVal)
			}

			if err != nil {
				NewResErr(err, err.Error(), http.StatusInternalServerError, w)
				return
			}

			if err := json.NewEncoder(w).Encode([]byte("hello")); err != nil {
				NewResErr(err, JSONEncodingError.Error(), http.StatusInternalServerError, w)
				return
			}
		}

	}
}

/**
 * Professional PQ handlers
 */

var getProProfile = func(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method now allowed.", 403)
	}
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	pro, err := fb.GetProfile(r.Context(), params["id"])
	if err != nil {
		NewResErr(err, "Error getting result for professional", http.StatusNotFound, w)
		return
	}
	if err := json.NewEncoder(w).Encode(pro); err != nil {
		NewResErr(err, "Error parsing to JSON", 503, w)
		return
	}
}

/**
 * Booking postgres
 */

//  GET /booking/{uid}
var getBookingsByUID = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		proUID := params["uid"]

		bookings, err := pq.GetBookingsByUID(r.Context(), proUID)
		if err != nil {
			NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		if err := json.NewEncoder(w).Encode(bookings); err != nil {
			NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}
	}
}

// POST /booking/{uid}
var createBooking = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		var req storage.Booking
		params := mux.Vars(r)
		uid := params["proUID"]
		log.Infoln(uid)

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			NewResErr(err, "Error reading body", http.StatusBadRequest, w)
			return
		}
		defer r.Body.Close()

		// * Is the date zero valued (i.e. missing or wrongly formatted)
		tb := timeutil.NewTime(*req.DateStart, *req.DateEnd)
		if err := tb.IsZero(); err != nil {
			NewResErr(err, err.Error(), http.StatusBadRequest, w, "trace")
			return
		}

		b, err := pq.CreateBooking(r.Context(), uid, req)
		if err != nil {
			NewResErr(err, err.Error(), http.StatusBadRequest, w, "trace")
			return
		}
		if err := json.NewEncoder(w).Encode(b); err != nil {
			NewResErr(err, err.Error(), http.StatusInternalServerError, w)
			return
		}
	}

}

// PUT /booking/{bookingID}
var updateBooking = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		w.Header().Set("Content-Type", "application/json")
		var b storage.Booking
		var err error
		params := mux.Vars(r)
		bookingID, ok := params["bookingID"]
		if !ok {
			err := fmt.Errorf("No booking ID provided")
			NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		b.ID = bookingID
		b.Task = r.FormValue("task")
		b.IsActive, err = conversion.ParseBool(r.FormValue("isActive"))
		if err != nil {
			NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		if err := pq.UpdateBooking(r.Context(), &b); err != nil {
			NewResErr(err, "Error inserting record", http.StatusInternalServerError, w, "trace")
			return
		}

		if err := json.NewEncoder(w).Encode(&b); err != nil {
			NewResErr(err, "Error returning response", http.StatusInternalServerError, w)
			return
		}
	}
}

// DELETE /bookings/{bookingID}
var deleteBooking = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		bookingID := params["bookingID"]
		if err := pq.DeleteBooking(r.Context(), bookingID); err != nil {
			NewResErr(err, "Error inserting record", http.StatusInternalServerError, w, "trace")
			return
		}
		if err := json.NewEncoder(w).Encode(&bookingID); err != nil {
			NewResErr(err, "Error sending response", http.StatusInternalServerError, w)
			return
		}
	}
}

// Gets the firebase profile, with postgres profile and booking
var getProfileWithBookings = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		profiles, err := pq.GetBookingsAdmin(r.Context())
		if err != nil {
			NewResErr(err, "Error getting value in database", http.StatusInternalServerError, w, "trace")
			return
		}
		for _, p := range profiles {
			fbprofile, err := fb.GetProfile(r.Context(), p.Professional.UserUID)
			if err != nil {
				NewResErr(err, "Error getting value in database", http.StatusInternalServerError, w, "trace")
				return
			}
			p.FirebaseProfile = *fbprofile
		}

		if err := json.NewEncoder(w).Encode(&profiles); err != nil {
			NewResErr(err, "Error sending response", http.StatusInternalServerError, w)
			return
		}
	}
}

// Response from byrd API OK/ERROR?
// var chargeBooking = func(w http.ResponseWriter, r *http.Request) {
// 	// TODO: get byrd api url to charge credits
// 	url := os.Getenv("ENV") + "/wht?"
// 	var client http.Client

// 	req, err := http.NewRequest("POST", url, r.Body)
// 	if err != nil {
// 		return
// 	}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return
// 	}

// 	if err := json.NewEncoder(w).Encode(res); err != nil {
// 		NewResErr(err, "Error encoding response", http.StatusInternalServerError, w, "trace")
// 		return
// 	}
// }

var sendMail = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("Content-type", "application/json")
		req := mail.RequestBody{}
		client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API"))
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Wrong body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		resp, err := req.SendMail(client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
