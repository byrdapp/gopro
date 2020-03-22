package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sendgrid/sendgrid-go"

	"github.com/byrdapp/byrd-pro-api/internal/mail"
	"github.com/byrdapp/byrd-pro-api/internal/storage"
	"github.com/byrdapp/byrd-pro-api/internal/storage/postgres"
	"github.com/byrdapp/byrd-pro-api/public/conversion"
	"github.com/byrdapp/byrd-pro-api/public/metadata"
	"github.com/byrdapp/byrd-pro-api/public/thumbnail"
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

func (s *server) loginGetUserAccess() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ? verify here, that the user is a pro user
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			var err error
			var creds Credentials
			if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}
			defer r.Body.Close()
			if creds.Password == "" || creds.Email == "" {
				s.writeClient(w, http.StatusBadRequest)
				return
			}

			usr, err := s.fb.GetProfileByEmail(r.Context(), creds.Email)
			if err != nil {
				s.writeClient(w, http.StatusNotFound)
				return
			}

			isPro, err := s.fb.IsProfessional(r.Context(), usr.UID)
			if !isPro || err != nil {
				s.writeClient(w, http.StatusForbidden)
				return
			}

			// Is user an admin? Set claims as such.
			// claims := make(map[string]interface{})
			isAdmin, err := s.fb.IsAdminUID(r.Context(), usr.UID)
			if err != nil {
				s.writeClient(w, http.StatusForbidden)
				return
			}
			credsRes := credsResponse{
				IsPro:   isPro,
				IsAdmin: isAdmin,
			}

			if err := json.NewEncoder(w).Encode(&credsRes); err != nil {
				s.writeClient(w, StatusJSONEncode)
				return
			}
		}
	}
}

// /profile/decode func attempts to return a profile from a given client UID header
func (s *server) decodeTokenGetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var err error
			clientToken := r.Header.Get(userToken)
			if clientToken == "" {
				s.writeClient(w, StatusBadTokenHeader)
				return
			}
			fbtoken, err := s.fb.VerifyToken(r.Context(), clientToken)
			if err != nil {
				s.writeClient(w, StatusBadTokenHeader)
				return
			}
			profile, err := s.fb.GetProfile(r.Context(), fbtoken.UID)
			if err != nil {
				s.writeClient(w, http.StatusNotFound)
				return
			}

			if err := json.NewEncoder(w).Encode(profile); err != nil {
				s.writeClient(w, StatusJSONEncode)
				return
			}
		}
	}
}

// /profile/{uid}
func (s *server) getProfileByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			params := mux.Vars(r)
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
			defer cancel()
			val, err := s.fb.GetProfile(ctx, params["id"])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			if err := json.NewEncoder(w).Encode(val); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

// getProfiles endpoint: /profiles
func (s *server) getProfiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("content-type", "application/json")
		medias, err := s.fb.GetProfiles(r.Context())
		if err != nil {
			s.writeClient(w, http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(medias); err != nil {
			s.writeClient(w, StatusJSONEncode)
		}
	}
}

// getExif receives body with img files
// it attempts to fetch EXIF data from each image
// if no exif data, the error message will be added to the response without breaking out of the loop until EOF.
// endpoint: exif/${type=image/video}/?preview:bool
func (s *server) exifImages() http.HandlerFunc {
	type response struct {
		Meta      *metadata.Metadata       `json:"meta,omitempty"`
		Thumbnail thumbnail.ImageThumbnail `json:"thumbnail,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var withPreview bool
			// w.Header().Set("Content-Type", "multipart/form-data")
			_, cancel := context.WithTimeout(r.Context(), time.Second*30)
			defer cancel()
			// Parse media type to get type of media
			mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}
			if !strings.HasPrefix(mediaType, "multipart/") {
				s.writeClient(w, StatusNotMultipart)
				return
			}

			withPreview = strings.EqualFold(r.URL.Query().Get("preview"), "true")
			var res []*response

			mr, err := r.MultipartReader()
			if err != nil {
				// panic(err)
			}
			defer r.Body.Close()
			for {
				part, err := mr.NextRawPart()
				if err != nil {
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						break
					}
					s.writeClient(w, http.StatusBadRequest)
					return
				}
				fileName := strings.ToLower(part.FileName())
				var data response
				if !metadata.SupportedImageSuffix(fileName) {
					s.writeClient(w, http.StatusUnsupportedMediaType)
					return
				}

				b, err := ioutil.ReadAll(part)
				if err != nil {

				}
				defer part.Close()

				br := bytes.NewReader(b)
				m, err := metadata.DecodeImage(br)
				if err != nil {
					s.Errorf("parsed exif error: %v on file: %v", err, fileName)
				}
				data.Meta = m

				if withPreview {
					if _, err := br.Seek(0, 0); err != nil {
						data.Thumbnail = nil
						continue
					}
					t := thumbnail.New(br)
					thumb, err := t.ImageThumbnail(300, 300)
					if err != nil {
						s.Warnf("thumbnail failed: %v", err)
					}
					data.Thumbnail = thumb
				}
				res = append(res, &data)
			}

			if err := json.NewEncoder(w).Encode(res); err != nil {
				s.writeClient(w, StatusJSONEncode)
				return
			}
		}
	}
}

func (s *server) exifVideo() http.HandlerFunc {
	type response struct {
		Meta      *metadata.Metadata        `json:"meta,omitempty"`
		Thumbnail thumbnail.FFMPEGThumbnail `json:"thumbnail,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, cancel := context.WithTimeout(r.Context(), time.Second*30)
		defer cancel()
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			s.writeClient(w, http.StatusBadRequest)
			return
		}

		// Wrong request header from filetype
		if !strings.HasPrefix(mediaType, "video/") {
			s.writeClient(w, http.StatusBadRequest)
			return
		}

		supported := metadata.SupportedVideoSuffix(mediaType)
		if !supported {
			s.writeClient(w, http.StatusUnsupportedMediaType)
			return
		}
		defer r.Body.Close()

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.writeClient(w, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		rd := bytes.NewReader(b)

		var res response

		meta, err := metadata.DecodeVideo(rd)
		if err != nil {
			s.writeClient(w, http.StatusBadRequest)
			return
		}
		res.Meta = meta
		// res.Size = conversion.FileSizeBytesToFloat(len(buf.Bytes()))

		if strings.EqualFold(r.URL.Query().Get("preview"), "true") {
			if _, err := rd.Seek(0, 0); err != nil {
				s.Warnf("seek error internally: %v", err)
			} else {
				t := thumbnail.New(rd)
				ffmpegThumb, err := t.VideoThumbnail(300, 300)
				if err != nil {
					s.Warnf("thumbnail failed: %v", err)
				}
				res.Thumbnail = ffmpegThumb
			}
		}

		if err := json.NewEncoder(w).Encode(&res); err != nil {
			s.writeClient(w, StatusJSONEncode)
			return
		}
	}
}

/**
 * Professional PQ handlers
 */

func (s *server) getProProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		pro, err := s.fb.GetProfile(r.Context(), params["id"])
		if err != nil {
			s.writeClient(w, http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(pro); err != nil {
			s.writeClient(w, StatusJSONEncode)
			return
		}
	}
}

/**
 * Booking postgres
 */

//  GET /booking/{uid}
func (s *server) getBookingsByUID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			params := mux.Vars(r)
			userId := params["uid"]
			bookings, err := s.pq.GetBookingsByMediaUID(r.Context(), userId)
			if err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}

			if err := json.NewEncoder(w).Encode(bookings); err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}
		}
	}
}

// POST /booking/task
func (s *server) createBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req postgres.CreateBookingParams
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				s.writeClient(w, http.StatusBadRequest).LogError(err)
				return
			}
			defer r.Body.Close()
			spew.Dump(req)

			// * if bad date
			if req.DateStart.IsZero() || req.DateEnd.IsZero() || req.DateEnd.Unix() < time.Now().Unix() {
				s.writeClient(w, StatusBadDateTime)
				return
			}

			req.Price = req.Credits * 15

			// ? minimum price cap??
			uuid, err := s.pq.CreateBooking(r.Context(), req)
			if err != nil {
				s.writeClient(w, http.StatusForbidden).LogError(err)
				return
			}

			if err := json.NewEncoder(w).Encode(uuid); err != nil {
				s.writeClient(w, http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *server) acceptBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

// PUT /booking/{bookingID}
func (s *server) updateBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			var b storage.Booking
			var err error
			params := mux.Vars(r)
			bookingID, ok := params["bookingID"]
			if !ok {
				s.writeClient(w, http.StatusBadRequest)
				return
			}

			b.ID = bookingID
			b.Task = r.FormValue("task")
			b.IsActive, err = conversion.ParseBool(r.FormValue("isActive"))
			if err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}

			// if err := pq.UpdateBooking(r.Context(), &b); err != nil {
			// 	s.WriteClient(w, http.StatusInternalServerError)
			// 	return
			// }

			if err := json.NewEncoder(w).Encode(&b); err != nil {
				s.writeClient(w, http.StatusInternalServerError)
				return
			}
		}
	}
}

// DELETE /bookings/{bookingID}
func (s *server) deleteBooking() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.Header().Set("Content-Type", "application/json")
			params := mux.Vars(r)
			userUID, err := uuid.FromBytes([]byte(params["bookingID"]))
			if err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}
			if err := s.pq.DeleteBooking(r.Context(), userUID); err != nil {
				s.writeClient(w, http.StatusInternalServerError)
				return
			}
			s.writeClient(w, http.StatusOK)
			return
		}
	}
}

func (s *server) sendMail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}
