package exif

import (
	"github.com/pkg/errors"
	goexif "github.com/rwcarlsen/goexif/exif"

	"github.com/blixenkrone/gopro/pkg/logger"
)

// type MediaDimension int

// Output represents the final decoded EXIF data from an image
type Output struct {
	// File            file.FileGenerator
	Date            int64             `json:"date,omitempty"`
	Lat             float64           `json:"lat,omitempty"`
	Lng             float64           `json:"lng,omitempty"`
	Copyright       string            `json:"copyright,omitempty"`
	Model           string            `json:"model,omitempty"`
	PixelXDimension int               `json:"pixelXDimension,omitempty"`
	PixelYDimension int               `json:"pixelYDimension,omitempty"`
	MediaSize       float64           `json:"mediaSize,omitempty"`
	MissingExif     map[string]string `json:"missingExif,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

var log = logger.NewLogger()

// adds an object to the output JSON that displays missing exif data
func (o *Output) AddMissingExif(errType string, originError error) {
	var returnError error
	if goexif.IsTagNotPresentError(originError) {
		returnError = errors.Wrapf(originError, "exif missing from type: %s", errType)
	}
	if goexif.IsCriticalError(originError) {
		returnError = errors.Wrapf(originError, "critical error with %s", errType)
	}
	if goexif.IsGPSError(originError) {
		returnError = errors.Wrapf(originError, "geographical error with %s", errType)
	}
	log.Errorf("Origin error: %s - Client error: %s - Type?: %s", originError, returnError, errType)
	returnError = errors.Errorf("error parsing from type %s", errType)
	o.MissingExif[errType] = returnError.Error()
}

func (o *Output) MediaType() (mediaType *string) {
	// TODO: handle this stuff
	switch o.Copyright {
	case "image":
		*mediaType = "image"
	case "video":
		*mediaType = "video"
	default:
		mediaType = nil
	}
	return mediaType
}
