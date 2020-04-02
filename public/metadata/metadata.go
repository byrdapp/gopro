package metadata

import (
	"errors"
	"io"

	"github.com/byrdapp/byrd-pro-api/public/conversion"
	"github.com/byrdapp/byrd-pro-api/public/metadata/image"
	"github.com/byrdapp/byrd-pro-api/public/metadata/video"
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

func DecodeVideo(r io.Reader) (*Metadata, error) {
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

	dateMillis := m.CreationTimeMillisUTC()
	if dateMillis.IsZero() || dateMillis.Unix() <= 0 {
		nilKeys = append(nilKeys, "date")
	}

	meta = Metadata{
		Date: dateMillis.Unix(),
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
func DecodeImage(r io.Reader) (*Metadata, error) {
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
		nilKeys = append(nilKeys, "geo")
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

	return &Metadata{
		Lat:       lat,
		Lng:       lng,
		Date:      date,
		Model:     model,
		Width:     w,
		Height:    h,
		Copyright: copyright,
		NilKeys:   nilKeys,
	}, nil
}
