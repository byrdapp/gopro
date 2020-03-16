package metadata

import (
	"errors"
	"fmt"
	"io"

	"github.com/byrdapp/byrd-pro-api/pkg/conversion"
	"github.com/byrdapp/byrd-pro-api/pkg/metadata/video"
	goexif "github.com/rwcarlsen/goexif/exif"
)

// Output represents the final decoded EXIF data from an image
type Metadata struct {
	// File            file.FileGenerator
	Date        int64             `json:"date,omitempty"`
	Lat         float64           `json:"lat,omitempty"`
	Lng         float64           `json:"lng,omitempty"`
	Copyright   string            `json:"copyright,omitempty"`
	Model       string            `json:"model,omitempty"`
	Height      int               `json:"height,omitempty"`
	Width       int               `json:"width,omitempty"`
	MediaSize   float64           `json:"mediaSize,omitempty"`
	MissingExif map[string]string `json:"missingExif,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

func VideoMetadata(r io.Reader) (*Metadata, error) {
	v := video.New(r)
	m, err := v.RawMeta()
	if err != nil {
		return nil, err
	}
	m.SanitizeOutput()

	meta := &Metadata{
		Date: m.CreationTime().UTC().Unix(),
		Lat:  conversion.MustStringToFloat(m.Lat()),
		Lng:  conversion.MustStringToFloat(m.Lng()),
		// Copyright: nil,
		Model:  m.Model(),
		Width:  m.Width(),
		Height: m.Height(),
	}
	return meta, nil
}

// adds an object to the output JSON that displays missing exif data
func (o *Metadata) AddMissingExif(tag string, originError error) {
	var returnErr error
	if goexif.IsTagNotPresentError(originError) {
		returnErr = errors.New("exif tag is not present in file")
	}
	if goexif.IsCriticalError(originError) {
		returnErr = fmt.Errorf("error parsing %v from file", tag)
	}
	if goexif.IsGPSError(originError) {
		returnErr = errors.New("GPS decoding error")
	}
	o.MissingExif[tag] = returnErr.Error()
}
