package exif

import (
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
	ExifErrors      map[string]string `json:"errors,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

var log = logger.NewLogger()

// adds an object to the output JSON that displays missing exif data
func (o *Output) MissingExif(errType string, err error) {
	log.Warnf("exif missing: %s from type: %s", err, errType)
	o.ExifErrors[errType] = err.Error()
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


