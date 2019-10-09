package exif

import (
	"fmt"
	"io"

	"github.com/blixenkrone/gopro/pkg/logger"

	"github.com/rwcarlsen/goexif/exif"
)

var (
	log = logger.NewLogger()
)

// tiff.Tag struct return values as number(i.e. 0 == int)
const (
	exifIntVal = iota
)

// Output represents the final decoded EXIF data from an image
type Output struct {
	Date            int64   `json:"date,omitempty"`
	Lat             float64 `json:"lat,omitempty"`
	Lng             float64 `json:"lng,omitempty"`
	Copyright       string  `json:"copyright,omitempty"`
	Model           string  `json:"model,omitempty"`
	PixelXDimension int     `json:"pixelXDimension,omitempty"`
	PixelYDimension int     `json:"pixelYDimension,omitempty"`
	MediaSize       int     `json:"mediaSize,omitempty"`
	MediaFormat     string  `json:"mediaFormat,omitempty"`
}

// GetOutput returns the struct *Output containing img data. Call this for each img.
func GetOutput(r io.Reader) (*Output, error) {
	x, err := loadExifData(r)
	if err != nil {
		return nil, fmt.Errorf("Error loading exif: %s", err)
	}
	lat, err := x.calcGeoCoordinate(exif.GPSLatitude)
	if err != nil {
		return nil, fmt.Errorf("Error getting lat data: %s", err)
	}
	lng, err := x.calcGeoCoordinate(exif.GPSLongitude)
	if err != nil {
		return nil, fmt.Errorf("Error getting lng data: %s", err)
	}
	date, err := x.getDateTime()
	if err != nil {
		return nil, fmt.Errorf("Error getting datetime: %s", err)
	}
	author, err := x.getCopyright()
	if err != nil {
		return nil, fmt.Errorf("Error getting copyright: %s", err)
	}
	model, err := x.getCameraModel()
	if err != nil {
		return nil, fmt.Errorf("Error getting camera model: %s", err)
	}

	fmtMap, err := x.getImageFormatData()
	if err != nil {
		return nil, fmt.Errorf("Error getting img fmt data: %s", err)
	}

	return &Output{
		Lat:             lat,
		Lng:             lng,
		Date:            date,
		Model:           model,
		PixelXDimension: fmtMap[exif.PixelXDimension],
		PixelYDimension: fmtMap[exif.PixelYDimension],
		Copyright:       author,
	}, nil
}

type imgExifData struct {
	x *exif.Exif
}

// loadExifData request exif data for image
func loadExifData(r io.Reader) (*imgExifData, error) {
	x, err := exif.Decode(r)
	if err != nil {
		log.Errorln("ERROR DECODING: " + err.Error())
		return nil, fmt.Errorf("Error decoding EXIF in image")
	}
	return &imgExifData{x}, nil
}

func (e *imgExifData) calcGeoCoordinate(fieldName exif.FieldName) (float64, error) {
	tag, err := e.x.Get(fieldName)
	if err != nil {
		if exif.IsTagNotPresentError(err) {
			log.Errorf("Error reading Geolocation in EXIF: %s", err)
			return 0.0, fmt.Errorf("Error reading Geolocation: %s", err)
		}
		return 0.0, err
	}
	ratVals := map[string]int{"deg": 0, "min": 1, "sec": 2}
	fVals := make(map[string]float64, len(ratVals))

	for key, val := range ratVals {
		rVals, err := tag.Rat(val)
		if err != nil {
			return 0.0, err
		}
		f, _ := rVals.Float64()
		fVals[key] = f
	}

	res := fVals["deg"] + (fVals["min"] / 60) + (fVals["sec"] / 3600)
	return res, nil
}

func (e *imgExifData) getDateTime() (d int64, err error) {
	t, err := e.x.DateTime()
	if err != nil {
		return d, err
	}
	d = t.UTC().Unix()
	return d, nil
}

func (e *imgExifData) getCopyright() (author string, err error) {
	tag, err := e.x.Get(exif.Copyright)
	if err != nil {
		return author, err
	}
	return tag.StringVal()
}

func (e *imgExifData) getCameraModel() (model string, err error) {
	n := exif.FieldName(exif.Model)
	tag, err := e.x.Get(n)
	if err != nil {
		return model, err
	}
	return tag.StringVal()
}

func (e *imgExifData) getImageFormatData() (map[exif.FieldName]int, error) {
	var fNames = []exif.FieldName{exif.PixelXDimension, exif.PixelYDimension}
	var fNameVal = make(map[exif.FieldName]int, len(fNames))
	for _, n := range fNames {
		tag, err := e.x.Get(n)
		if err != nil {
			return nil, err
		}
		i, err := tag.Int(exifIntVal)
		if err != nil {
			return nil, err
		}
		fNameVal[n] = i
	}
	return fNameVal, nil
}

// func (e *imgExifData) getMediaSize() (int, error) {
// 	e.x.Get(exif)
// }

// ? not in use currently
func (e *imgExifData) getExifFieldNameString(fieldName exif.FieldName) (string, error) {
	tag, _ := e.x.Get(fieldName)
	return tag.StringVal()
}
