package metadata

import (
	"errors"
	"io"

	"github.com/byrdapp/byrd-pro-api/pkg/conversion"
	"github.com/byrdapp/byrd-pro-api/pkg/metadata/image"
	"github.com/byrdapp/byrd-pro-api/pkg/metadata/video"
)

var (
	EOFError = errors.New("error reading exif from file")
)

// Output represents the final decoded EXIF data from an image
type Metadata struct {
	// File            file.FileGenerator
	Date      int64    `json:"date,omitempty"`
	Lat       float64  `json:"lat,omitempty"`
	Lng       float64  `json:"lng,omitempty"`
	Copyright string   `json:"copyright,omitempty"`
	Model     string   `json:"model,omitempty"`
	Height    int      `json:"height,omitempty"`
	Width     int      `json:"width,omitempty"`
	MediaSize float64  `json:"mediaSize,omitempty"`
	NilKeys   []string `json:"missingExif,omitempty"`
	// MissingExif map[string]string `json:"missingExif,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

func VideoMetadata(r io.Reader) (*Metadata, error) {
	v := video.New(r)
	m, err := v.RawMeta()
	if err != nil {
		return nil, err
	}
	m.SanitizeOutput()

	var meta Metadata
	var nilKeys []string

	model, err := m.Model()
	if err != nil {
		nilKeys = append(nilKeys, "model")
	}

	if m.CreationTime().UTC().IsZero() || m.CreationTime().UTC().Unix() < 0 {
		nilKeys = append(nilKeys, "date")
	}

	meta = Metadata{
		Date: m.CreationTime().UTC().Unix(),
		Lat:  conversion.MustStringToFloat(m.Lat()),
		Lng:  conversion.MustStringToFloat(m.Lng()),
		// Copyright: nil,
		Model:   model,
		Width:   m.Width(),
		Height:  m.Height(),
		NilKeys: nilKeys,
	}
	return &meta, nil
}

// DecodeImageMetadata returns the struct *Output containing img data.
// This will include the errors from missing/broken exif will follow.
// If an error is != nil, its a panic
func DecodeImageMetadata(r io.Reader) (*Metadata, error) {
	// r := bytes.NewReader(data)
	// xErr := &metadata.Metadata{MissingExif: make(map[string]string)}
	var nilKeys []string
	m, err := image.ImageMetadata(r)
	if err != nil {
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			return nil, EOFError
		}
		// Missing exif should probably not happen
		return nil, errors.New("error decoding image for meta data")
	}
	lat, lng, err := m.Geo()
	if err != nil {
		nilKeys = append(nilKeys, "lat")
	}
	date, err := m.DateMillisUnix()
	if err != nil {
		nilKeys = append(nilKeys, "date")
	}
	copyright, err := m.Copyright()
	if err != nil {
		nilKeys = append(nilKeys, "copyright")
	}
	model, err := m.Model()
	if err != nil {
		nilKeys = append(nilKeys, "model")
	}
	w, h, err := m.Dimensions()
	if err != nil {
		nilKeys = append(nilKeys, "dimension")
	}
	// ! if err != nil {
	// 	size, err := m.getFileSize(r)
	nilKeys = append(nilKeys, "size")
	// }

	return &Metadata{
		Lat:       lat,
		Lng:       lng,
		Date:      date,
		Model:     model,
		Width:     w,
		Height:    h,
		Copyright: copyright,
		// MediaSize: size,
		NilKeys: nilKeys,
		// ? do this MediaFormat:     mediaFmt,
	}, nil
}

// // adds an object to the output JSON that displays missing exif data
// func (o *Metadata) AddMissingExif(tag string, originError error) {
// 	var returnErr error
// 	if goexif.IsTagNotPresentError(originError) {
// 		returnErr = errors.New("exif tag is not present in file")
// 	}
// 	if goexif.IsCriticalError(originError) {
// 		returnErr = fmt.Errorf("error parsing %v from file", tag)
// 	}
// 	if goexif.IsGPSError(originError) {
// 		returnErr = errors.New("GPS decoding error")
// 	}
// 	_ = returnErr
// 	//!  o.MissingExif[tag] = returnErr.Error()
// }
