package media

import (
	"errors"
	"fmt"

	goexif "github.com/rwcarlsen/goexif/exif"
)

// type MediaDimension int

// Output represents the final decoded EXIF data from an image
type Metadata struct {
	// File            file.FileGenerator
	Date            int64             `json:"date,omitempty"`
	Lat             float64           `json:"lat,omitempty"`
	Lng             float64           `json:"lng,omitempty"`
	ISOLat          float64           `json:"isoLat,omitempty"`
	ISOLng          float64           `json:"isoLng,omitempty"`
	Copyright       string            `json:"copyright,omitempty"`
	Model           string            `json:"model,omitempty"`
	PixelXDimension int               `json:"pixelXDimension,omitempty"`
	PixelYDimension int               `json:"pixelYDimension,omitempty"`
	MediaSize       float64           `json:"mediaSize,omitempty"`
	MissingExif     map[string]string `json:"missingExif,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
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
