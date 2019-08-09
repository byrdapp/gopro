package main

import (
	"encoding/json"
	stdliberr "errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/blixenkrone/gopro/securion"
	postgres "github.com/blixenkrone/gopro/storage/postgres"
	exif "github.com/blixenkrone/gopro/upload/exif"
	"github.com/blixenkrone/gopro/utils/errors"
	mux "github.com/gorilla/mux"
	goexif "github.com/rwcarlsen/goexif/exif"
	"golang.org/x/net/context"
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
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			errors.NewResErr(err, "Error decoding JSON from request body", http.StatusBadRequest, w)
			return
		}
		defer r.Body.Close()

		if creds.Password == "" || creds.Email == "" {
			err := stdliberr.New("Missing email or password in credentials")
			errors.NewResErr(err, err.Error(), http.StatusInternalServerError, w)
			return
		}

		usr, err := fb.GetProfileByEmail(r.Context(), creds.Email)
		if err != nil {
			errors.NewResErr(err, "Error finding profile UID in Firebase Auth. Does the user exist?", http.StatusGone, w)
			return
		}

		signedToken, err := fb.CreateCustomToken(r.Context(), usr.UID)
		if err != nil {
			spew.Errorf("Error: %s", err)
		}

		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			errors.NewResErr(err, "Error encoding JSON token", http.StatusInternalServerError, w)
			return
		}
	}
}

var decodeTokenGetProfile = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tkn := r.Header.Get(userToken)
		fbtoken, err := fb.VerifyToken(r.Context(), tkn)
		if err != nil {
			errors.NewResErr(err, "No token provided in headers", http.StatusBadRequest, w)
			return
		}
		val, err := fb.GetProfile(r.Context(), fbtoken.UID)
		if err != nil {
			errors.NewResErr(err, "Error getting profile", http.StatusInternalServerError, w)
			return
		}
		if err := json.NewEncoder(w).Encode(val); err != nil {
			errors.NewResErr(err, "Error encoding JSON token", http.StatusInternalServerError, w)
			return
		}

	}
}

var getMediaByID = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		params := mux.Vars(r)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		val, err := pq.GetMediaByID(ctx, params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err := json.NewEncoder(w).Encode(val); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Dont use this for public routes
var createMedia = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.Header.Set("content-type", "application/json")
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		var media postgres.Media
		if err := json.NewDecoder(r.Body).Decode(&media); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		defer r.Body.Close()

		id, err := pq.CreateMedia(ctx, &media)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = json.NewEncoder(w).Encode(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// getMedias endpoint: /medias
var getMedias = func(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("content-type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()
	// todo: params
	medias, err := pq.GetMedias(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := json.NewEncoder(w).Encode(medias); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TagResult struct for exif handler to return either result or err
type TagResult struct {
	Out *goexif.Exif `json:"exif,omitempty"`
	Lng float64      `json:"lng,omitempty"`
	Lat float64      `json:"lat,omitempty"`
	Err string       `json:"err,omitempty"`
}

var getSecurionPlans = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	var res []*securion.Plan
	interval := r.FormValue("interval")

	secClient := securion.NewClient()
	plans, err := secClient.GetPlansJSON("10", interval)
	if err != nil {
		errors.NewResErr(err, err.Error(), 503, w)
		return
	}
	log.Info(interval)
	for _, p := range plans {
		if p.Interval == interval {
			for _, std := range securion.StdPlans {
				if c := strings.Compare(p.ID, std); c == 0 {
					log.Info(p.ID)
					res = append(res, p)
				}
			}
		}
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		errors.NewResErr(err, err.Error(), 503, w)
		return
	}
}

// getExif recieves body with img files
var getExif = func(w http.ResponseWriter, r *http.Request) {
	// r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "multipart/form-data")
		defer r.Body.Close()
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

var getProStats = func(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method now allowed.", 403)
	}
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	stats, err := pq.GetProStats(ctx, params["id"])
	if err != nil {
		errors.NewResErr(err, "Error getting pro stats", 503, w)
		return
	}
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		errors.NewResErr(err, "Error parsing to JSON", 503, w)
		return
	}
}
