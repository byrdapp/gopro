package exif

import (
	"fmt"
	"io"

	"github.com/blixenkrone/gopro/utils/logger"

	"github.com/rwcarlsen/goexif/exif"
)

var (
	log = logger.NewLogger()
)

// Output represents the final decoded EXIF data from an image
type Output struct {
	Date      string
	Lng       float64
	Lat       float64
	Copyright string
}

// GetOutput returns the struct *Output containing img data. Call this for each img.
func GetOutput(r io.Reader) (*Output, error) {
	x, err := loadExifData(r)
	lat, err := x.calcGeoCoordinate(exif.GPSLatitude)
	if err != nil {
		return nil, err
	}
	lng, err := x.calcGeoCoordinate(exif.GPSLongitude)
	if err != nil {
		return nil, err
	}
	date, err := x.getDateTime()
	if err != nil {
		return nil, err
	}
	author, err := x.getCopyright()
	res := &Output{
		Lat:       lat,
		Lng:       lng,
		Date:      date,
		Copyright: author,
	}
	return res, nil
}

type imgExifData struct {
	x *exif.Exif
}

// loadExifData request exif data for image
func loadExifData(r io.Reader) (*imgExifData, error) {
	x, err := exif.Decode(r)
	if err != nil {
		log.Errorln(err)
	}
	return &imgExifData{x}, nil
}

func (e *imgExifData) calcGeoCoordinate(fieldName exif.FieldName) (float64, error) {
	tag, err := e.x.Get(fieldName)
	if err != nil {
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

func (e *imgExifData) getDateTime() (date string, err error) {
	tag, err := e.x.Get(exif.DateTimeOriginal)
	if err != nil {
		return date, fmt.Errorf("Date error: %s", err)
	}
	date, err = tag.StringVal()
	if err != nil {
		return date, err
	}
	return date, nil
}

func (e *imgExifData) getCopyright() (author string, err error) {
	tag, err := e.x.Get(exif.Copyright)
	if err != nil {
		return author, err
	}
	return tag.StringVal()
}
