package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
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
	"github.com/byrdapp/byrd-pro-api/pkg/conversion"
	"github.com/byrdapp/byrd-pro-api/pkg/metadata"
	"github.com/byrdapp/byrd-pro-api/pkg/thumbnail"
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
				s.writeClient(w, StatusJSONDecode)
				return
			}
			defer r.Body.Close()
			if creds.Password == "" || creds.Email == "" {
				s.writeClient(w, http.StatusBadRequest)
				return
			}

			usr, err := s.fb.GetProfileByEmail(r.Context(), creds.Email)
			if err != nil {
				s.writeClient(w, http.StatusBadRequest)
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
				s.writeClient(w, http.StatusNotFound)
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

// response struct - dont use data as pointers, refer writer to the Metadata pointer when allocated.
type Metadata struct {
	Preview preview    `json:"preview,omitempty"`
	Exif    exifOutput `json:"exif,omitempty"`
}

type exifOutput struct {
	Output *metadata.Metadata `json:"output,omitempty"`
	Error  string             `json:"error,omitempty"`
}

type preview struct {
	Source []byte `json:"source,omitempty"`
	Error  string `json:"error,omitempty"`
}

// getExif receives body with img files
// it attempts to fetch EXIF data from each image
// if no exif data, the error message will be added to the response without breaking out of the loop until EOF.
// endpoint: exif/${type=image/video}/?preview:bool
func (s *server) exifImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.Body = http.MaxBytesReader(w, r.Body, 32<<20+512)
		if r.Method == http.MethodPost {
			var withPreview bool
			// w.Header().Set("Content-Type", "multipart/form-data")
			_, cancel := context.WithTimeout(r.Context(), time.Second*10)
			defer cancel()
			// Parse media type to get type of media
			mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil {
				s.writeClient(w, http.StatusBadRequest)
				return
			}
			if strings.HasPrefix(mediaType, "multipart/") {
				withPreview = strings.EqualFold(r.URL.Query().Get("preview"), "true")
				mr := multipart.NewReader(r.Body, params["boundary"])
				defer r.Body.Close()
				var res []*Metadata

				for {
					// (*os.File) for next file
					part, err := mr.NextPart()
					if err != nil {
						if err == io.EOF {
							break
						}
						s.writeClient(w, http.StatusNotAcceptable)
						return
					}

					var buf bytes.Buffer
					_, err = io.Copy(&buf, part)
					if err != nil {
						s.writeClient(w, http.StatusNotAcceptable)
						break
					}

					// JSON response struct
					var data Metadata
					if withPreview {
						// img, err := thumbnail.New(buf.Bytes())
						// if err != nil {
						// 	s.Warnf("%v", err)
						// 	data.Preview.Error = err.Error()
						// }
						// thumb, err := img.EncodeThumbnail()
						// if err != nil {
						// 	s.Errorf("%v", err)
						// 	s.WriteClient(w, http.StatusBadRequest)
						// 	return
						// }
						// data.Preview.Source = thumb.Bytes()
					}

					// Read EXIF data
					// parsedExif, err := image.DecodeImageMetadata(buf.Bytes())
					// if err != nil {
					// 	s.Errorf("parsed exif error: %v", err)
					// 	data.Exif.Error = err.Error()
					// }
					// data.Exif.Output = parsedExif
					res = append(res, &data)
				}

				if err := json.NewEncoder(w).Encode(res); err != nil {
					s.writeClient(w, StatusJSONEncode)
					return
				}
			}
		}
	}
}

func (s *server) exifVideo() http.HandlerFunc {
	type response struct {
		Meta      *metadata.Metadata        `json:"meta,omitempty"`
		Size      float64                   `json:"size,omitempty"`
		Thumbnail thumbnail.FFMPEGThumbnail `json:"thumbnail"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			s.writeClient(w, http.StatusUnsupportedMediaType)
			return
		}

		// Wrong request header from filetype
		if !strings.HasPrefix(mediaType, "multipart/") && !strings.HasPrefix(mediaType, "video/") {
			s.writeClient(w, http.StatusUnsupportedMediaType)
			return
		}

		supported := metadata.SupportedVideoSuffix(mediaType)
		if !supported {
			s.writeClient(w, http.StatusUnsupportedMediaType)
			return
		}

		var res response
		var buf bytes.Buffer
		tr := io.TeeReader(r.Body, &buf)
		defer r.Body.Close()
		res.Size = conversion.FileSizeBytesToFloat(len(buf.Bytes()))
		meta, err := metadata.VideoMetadata(tr)
		if err != nil {
			s.writeClient(w, http.StatusBadRequest)
			return
		}
		res.Meta = meta
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			s.writeClient(w, StatusJSONEncode)
			return
		}

	}
}

// /exif/video
func (s *server) exifVideoOld() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			var withPreview bool
			mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil {
				s.writeClient(w, http.StatusUnsupportedMediaType)
				return
			}

			// Wrong request body
			if !strings.HasPrefix(mediaType, "multipart/") && !strings.HasPrefix(mediaType, "video/") {
				s.writeClient(w, http.StatusUnsupportedMediaType)
				return
			}

			file, fheader, err := r.FormFile("file")
			if err != nil {
				s.writeClient(w, http.StatusForbidden)
				return
			}
			defer file.Close()

			filetype, ok := fheader.Header["Content-Type"]
			if !ok || len(filetype) == 0 {
				s.writeClient(w, http.StatusUnsupportedMediaType).LogError(err)
				return
			}
			var res Metadata
			// headerMediaType := strings.Split(filetype[0], "video/")[1]
			// fmt, err := media.Format(headerMediaType).Video()
			// if err != nil {
			// 	s.WriteClient(w, http.StatusUnsupportedMediaType).LogError(err)
			// 	return
			// }
			withPreview = strings.EqualFold(r.URL.Query().Get("preview"), "true")
			if withPreview {
				// t, err := video.Thumbnail(r.Body, 300, 300)
				// if err != nil {
				// 	s.Warnf("%v", err)
				// 	res.Preview.Error = err.Error()
				// }
				// res.Preview.Source = t
			}

			// out := video.Metadata()
			// res.Exif.Output = out

			if err := json.NewEncoder(w).Encode(&res); err != nil {
				s.writeClient(w, http.StatusInternalServerError)
				return
			}
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
