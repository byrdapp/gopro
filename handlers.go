package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	utils "github.com/blixenkrone/gopro/utils/fmt"

	mux "github.com/gorilla/mux"
	goexif "github.com/rwcarlsen/goexif/exif"

	storage "github.com/blixenkrone/gopro/storage"
	exif "github.com/blixenkrone/gopro/upload/exif"
	"github.com/blixenkrone/gopro/utils/errors"
	timeutil "github.com/blixenkrone/gopro/utils/time"
)

/**
 * ! implement context in all server>db calls
 *
 * * New plan:
 * In PQ:
 * - Keep table "profile" or "professional" and have only UID from FB in there (at least not conflicting data from FB)
 * - refer to the UID when adding / getting data from PQ db
 */

var signOut = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "application/json")
		http.SetCookie(w, &http.Cookie{
			Name:   "pro_token",
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
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			errors.NewResErr(err, "Error decoding JSON from request body", http.StatusBadRequest, w)
			return
		}
		defer r.Body.Close()

		if creds.Password == "" || creds.Email == "" {
			err := fmt.Errorf("Missing email or password in credentials")
			errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
			return
		}

		usr, err := fb.GetProfileByEmail(r.Context(), creds.Email)
		if err != nil {
			errors.NewResErr(err, "Error finding profile UID in Firebase Auth. Does the user exist?", http.StatusGone, w)
			return
		}

		// Is user an admin? Set claims as such.
		claims := make(map[string]interface{})
		isAdmin, err := fb.IsAdminUID(r.Context(), usr.UID)
		if err != nil {
			errors.NewResErr(err, "Error admin ref was not found", http.StatusGone, w)
			return
		}
		claims[isAdminClaim] = isAdmin
		signedToken, err := fb.CreateCustomTokenWithClaims(r.Context(), usr.UID, claims)
		if err != nil {
			errors.NewResErr(err, "Error creating token!", http.StatusInternalServerError, w, "trace")
			return
		}

		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			errors.NewResErr(err, "Error encoding JSON token", http.StatusInternalServerError, w)
			return
		}
	}
}

// /profile/decode func attempts to return a profile from a given client UID header
var decodeTokenGetProfile = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var err error
		clientToken := r.Header.Get(userToken)
		defer r.Body.Close()
		if clientToken == "" {
			err = fmt.Errorf("No header token from client")
			errors.NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}
		fbtoken, err := fb.VerifyToken(r.Context(), clientToken)
		if err != nil {
			errors.NewResErr(err, "No token provided in headers", http.StatusBadRequest, w)
			return
		}
		profile, err := fb.GetProfile(r.Context(), fbtoken.UID)
		if err != nil {
			errors.NewResErr(err, "Error getting profile", http.StatusInternalServerError, w)
			return
		}
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			errors.NewResErr(err, "Error encoding JSON token", http.StatusInternalServerError, w)
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
		errors.NewResErr(err, "Error finding media profiles", http.StatusFound, w)
	}
	if err := json.NewEncoder(w).Encode(medias); err != nil {
		errors.NewResErr(err, "JSON Encoding failed", 500, w)
	}
}

// TagResult struct for exif handler to return either result or err
type TagResult struct {
	Out *goexif.Exif `json:"exif,omitempty"`
	Lng float64      `json:"lng,omitempty"`
	Lat float64      `json:"lat,omitempty"`
	Err string       `json:"err,omitempty"`
}

// getExif recieves body with img files
var getExif = func(w http.ResponseWriter, r *http.Request) {
	// r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "multipart/form-data")
		_, cancel := context.WithTimeout(r.Context(), time.Duration(time.Second*10))
		defer cancel()

		// Parse media type to get type of media
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			errors.NewResErr(err, "Could not parse request body", http.StatusBadRequest, w)
			return
		}

		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(r.Body, params["boundary"])
			defer r.Body.Close()
			var exifRes []*TagResult
			for {
				var res TagResult
				part, err := mr.NextPart()
				// read length of files
				if err == io.EOF {
					log.Infoln("No more files to read")
					break
				}
				if err != nil {
					errors.NewResErr(err, "Could not read file"+part.FileName(), http.StatusBadRequest, w)
					break
				}

				imgsrv, err := exif.NewExifReq(part)
				if err != nil {
					errors.NewResErr(err, err.Error(), 503, w)
					break
				}

				out, err := imgsrv.TagExifSync()
				if err != nil {
					res.Err = err.Error()
				} else {
					res.Out = out
					// res.Lat = lat
					// res.Lng = lng
				}
				exifRes = append(exifRes, &res)
			}
			if err := json.NewEncoder(w).Encode(exifRes); err != nil {
				errors.NewResErr(err, "Error convert exif to JSON", 503, w)
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
		errors.NewResErr(err, "Error getting result for professional", http.StatusNotFound, w)
		return
	}
	if err := json.NewEncoder(w).Encode(pro); err != nil {
		errors.NewResErr(err, "Error parsing to JSON", 503, w)
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

		bookings, err := pq.GetBookings(r.Context(), proUID)
		if err != nil {
			errors.NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		if err := json.NewEncoder(w).Encode(bookings); err != nil {
			errors.NewResErr(err, err.Error(), http.StatusBadRequest, w)
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
			errors.NewResErr(err, "Error reading body", http.StatusBadRequest, w)
			return
		}
		defer r.Body.Close()

		// * Is the date zero valued (i.e. missing or wrongly formatted)
		tb := timeutil.NewTime(req.DateStart, req.DateEnd)
		if err := tb.IsZero(); err != nil {
			errors.NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		b, err := pq.CreateBooking(r.Context(), uid, req)
		if err != nil {
			errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
			return
		}
		if err := json.NewEncoder(w).Encode(b); err != nil {
			errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
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
			errors.NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		b.ID = bookingID
		b.Task = r.FormValue("task")
		b.IsActive, err = utils.ParseBool(r.FormValue("isActive"))
		if err != nil {
			errors.NewResErr(err, err.Error(), http.StatusBadRequest, w)
			return
		}

		if err := pq.UpdateBooking(r.Context(), &b); err != nil {
			errors.NewResErr(err, "Error inserting record", http.StatusInternalServerError, w, "trace")
			return
		}

		if err := json.NewEncoder(w).Encode(&b); err != nil {
			errors.NewResErr(err, "Error returning response", http.StatusInternalServerError, w)
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
			errors.NewResErr(err, "Error inserting record", http.StatusInternalServerError, w, "trace")
			return
		}
		if err := json.NewEncoder(w).Encode(&bookingID); err != nil {
			errors.NewResErr(err, "Error sending response", http.StatusInternalServerError, w)
			return
		}
	}
}

var getBookingsAdmin = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res, err := pq.GetBookingsAdmin(r.Context())
		if err != nil {
			errors.NewResErr(err, "Error getting value in database", http.StatusInternalServerError, w, "trace")
			return
		}

		if err := json.NewEncoder(w).Encode(&res); err != nil {
			errors.NewResErr(err, "Error sending response", http.StatusInternalServerError, w)
			return
		}
	}
}

// Response from byrd API OK/ERROR?
var chargeBooking = func(w http.ResponseWriter, r *http.Request) {}
