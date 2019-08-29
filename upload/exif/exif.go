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
	Date      string  `json:"date,omitempty"`
	Lng       float64 `json:"lng,omitempty"`
	Lat       float64 `json:"lat,omitempty"`
	Copyright string  `json:"copyright,omitempty"`
}

// Error is error struct to the client
type Error struct {
	Message string `json:"msg,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// GetOutput returns the struct *Output containing img data. Call this for each img.
func GetOutput(r io.Reader) (*Output, error) {
	x, err := loadExifData(r)
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
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
		log.Errorln("ERROR DECODING:" + err.Error())
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
