package server

import (
	"bytes"
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
	"github.com/sendgrid/sendgrid-go"

	"github.com/blixenkrone/gopro/internal/mail"
	"github.com/blixenkrone/gopro/internal/storage"
	"github.com/blixenkrone/gopro/pkg/conversion"
	"github.com/blixenkrone/gopro/pkg/image/thumbnail"
	"github.com/blixenkrone/gopro/pkg/media"
	image "github.com/blixenkrone/gopro/pkg/media/image"
	video "github.com/blixenkrone/gopro/pkg/media/video"
	timeutil "github.com/blixenkrone/gopro/pkg/time"
)

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

type credsResponse struct {
	IsPro   bool `json:"isPro"`
	IsAdmin bool `json:"isAdmin"`
}

var loginGetUserAccess = func(w http.ResponseWriter, r *http.Request) {
	// ? verify here, that the user is a pro user
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		var err error
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			WriteClient(w, StatusJSONDecode)
			return
		}
		defer r.Body.Close()
		if creds.Password == "" || creds.Email == "" {
			WriteClient(w, http.StatusBadRequest)
			return
		}

		usr, err := fb.GetProfileByEmail(r.Context(), creds.Email)
		if err != nil {
			WriteClient(w, http.StatusBadRequest)
			return
		}

		isPro, err := fb.IsProfessional(r.Context(), usr.UID)
		if !isPro || err != nil {
			WriteClient(w, http.StatusForbidden)
			return
		}

		// Is user an admin? Set claims as such.
		// claims := make(map[string]interface{})
		isAdmin, err := fb.IsAdminUID(r.Context(), usr.UID)
		if err != nil {
			WriteClient(w, http.StatusNotFound)
			return
		}
		credsRes := credsResponse{
			IsPro:   isPro,
			IsAdmin: isAdmin,
		}

		if err := json.NewEncoder(w).Encode(&credsRes); err != nil {
			WriteClient(w, StatusJSONEncode)
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
			WriteClient(w, StatusBadTokenHeader)
			return
		}
		fbtoken, err := fb.VerifyToken(r.Context(), clientToken)
		if err != nil {
			WriteClient(w, StatusBadTokenHeader)
			return
		}
		profile, err := fb.GetProfile(r.Context(), fbtoken.UID)
		if err != nil {
			WriteClient(w, http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(profile); err != nil {
			WriteClient(w, StatusJSONEncode)
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
		WriteClient(w, http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(medias); err != nil {
		WriteClient(w, StatusJSONEncode)
	}
}

type exifImagesResponse struct {
	Preview *preview    `json:"preview,omitempty"`
	Exif    *exifOutput `json:"exif,omitempty"`
}

type exifOutput struct {
	Output *media.Metadata `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type preview struct {
	Source []byte `json:"source,omitempty"`
	Error  string `json:"error,omitempty"`
}

// getExif receives body with img files
// it attempts to fetch EXIF data from each image
// if no exif data, the error message will be added to the response without breaking out of the loop until EOF.
// endpoint: exif/${type=image/video}/?preview:bool
var exifImages = func(w http.ResponseWriter, r *http.Request) {
	// r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)
	if r.Method == http.MethodPost {
		var withPreview bool
		// w.Header().Set("Content-Type", "multipart/form-data")
		_, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		// Parse media type to get type of media
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			WriteClient(w, http.StatusBadRequest)
			return
		}
		if strings.HasPrefix(mediaType, "multipart/") {
			withPreview = strings.EqualFold(r.URL.Query().Get("preview"), "true")
			log.Infof("withpreview: %v", withPreview)
			mr := multipart.NewReader(r.Body, params["boundary"])
			defer r.Body.Close()
			var res []*exifImagesResponse

			for {
				// (*os.File) for next file
				part, err := mr.NextPart()
				if err != nil {
					if err == io.EOF {
						break
					}
					WriteClient(w, http.StatusNotAcceptable).LogError(fmt.Errorf("file: %s + %+v", part.FileName(), err))
					break
				}

				var buf bytes.Buffer
				_, err = io.Copy(&buf, part)
				if err != nil {
					WriteClient(w, http.StatusNotAcceptable).LogError(fmt.Errorf("file: %s + %+v", part.FileName(), err))
					break
				}

				log.Infof("copied file: ", part.FileName())

				// JSON response struct
				var data exifImagesResponse

				if withPreview {
					var preview preview
					img, err := thumbnail.New(buf.Bytes())
					if err != nil {
						preview.Error = err.Error()
						log.Error(err)
					}
					thumb, err := img.EncodeThumbnail()
					if err != nil {
						WriteClient(w, http.StatusBadRequest)
						return
					}
					preview.Source = thumb.Bytes()
					data.Preview = &preview
				}

				// Read EXIF data
				var exif exifOutput
				parsedExif, err := image.DecodeImageMetadata(buf.Bytes())
				if err != nil {
					log.Errorf("parsed exif error: %v", err)
					exif.Error = err.Error()
				}
				exif.Output = parsedExif
				data.Exif = &exif

				res = append(res, &data)
			}

			if err := json.NewEncoder(w).Encode(res); err != nil {
				WriteClient(w, StatusJSONEncode)
				return
			}
		}
	}
}

type videoReponse struct {
	Out       *media.Metadata `json:"out,omitempty"`
	Thumbnail []byte          `json:"thumbnail,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// /exif/video
var exifVideo = func(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		var withPreview bool
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			WriteClient(w, http.StatusUnsupportedMediaType)
			return
		}
		// Wrong request body

		log.Info(mediaType)
		if !strings.HasPrefix(mediaType, "multipart/") {
			WriteClient(w, http.StatusBadRequest)
			return
		}

		file, fheader, err := r.FormFile("file")
		if err != nil {
			WriteClient(w, http.StatusBadRequest)
			return
		}
		defer file.Close()

		headerMediaType := strings.Split(fheader.Header["Content-Type"][0], "video/")[1]
		fmt, ok, err := media.IsSupportedMediaFmt(headerMediaType)
		if !ok || err != nil {
			WriteClient(w, http.StatusUnsupportedMediaType).LogError(err)
			return
		}
		withPreview = strings.EqualFold(r.URL.Query().Get("preview"), "true")

		video, err := video.ReadVideoBuffer(file, fmt)
		if err != nil {
			WriteClient(w, http.StatusNotAcceptable)
			return
		}
		defer r.Body.Close()

		var res videoReponse
		if withPreview {
			t, err := video.Thumbnail()
			if err != nil {
				log.Warn(err)
				res.Error = err.Error()
			}
			res.Thumbnail = t.Bytes()
		}
		out := video.Metadata()
		res.Out = out

		if err := json.NewEncoder(w).Encode(&res); err != nil {
			WriteClient(w, http.StatusInternalServerError)
			return
		}
	}
}

/**
 * Professional PQ handlers
 */

var getProProfile = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	pro, err := fb.GetProfile(r.Context(), params["id"])
	if err != nil {
		WriteClient(w, http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(pro); err != nil {
		WriteClient(w, StatusJSONEncode)
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
			WriteClient(w, http.StatusBadRequest)
			return
		}

		if err := json.NewEncoder(w).Encode(bookings); err != nil {
			WriteClient(w, http.StatusBadRequest)
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
			WriteClient(w, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// * Is the date zero valued (i.e. missing or wrongly formatted)
		tb := timeutil.NewTime(*req.DateStart, *req.DateEnd)
		if err := tb.IsZero(); err != nil {
			WriteClient(w, http.StatusBadRequest)
			return
		}

		b, err := pq.CreateBooking(r.Context(), uid, req)
		if err != nil {
			WriteClient(w, http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(b); err != nil {
			WriteClient(w, http.StatusInternalServerError)
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
			WriteClient(w, http.StatusBadRequest)
			return
		}

		b.ID = bookingID
		b.Task = r.FormValue("task")
		b.IsActive, err = conversion.ParseBool(r.FormValue("isActive"))
		if err != nil {
			WriteClient(w, http.StatusBadRequest)
			return
		}

		if err := pq.UpdateBooking(r.Context(), &b); err != nil {
			WriteClient(w, http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(&b); err != nil {
			WriteClient(w, http.StatusInternalServerError)
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
			WriteClient(w, http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(&bookingID); err != nil {
			WriteClient(w, http.StatusInternalServerError)
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
			WriteClient(w, http.StatusInternalServerError)
			return
		}
		for _, p := range profiles {
			fbprofile, err := fb.GetProfile(r.Context(), p.Professional.UserUID)
			if err != nil {
				WriteClient(w, http.StatusInternalServerError)
				return
			}
			p.FirebaseProfile = *fbprofile
		}

		if err := json.NewEncoder(w).Encode(&profiles); err != nil {
			WriteClient(w, http.StatusInternalServerError)
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
// 		JSONWrite(w, "Error encoding response", http.StatusInternalServerError, w, "trace")
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
