package exif_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/blixenkrone/gopro/utils/logger"

	"github.com/davecgh/go-spew/spew"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

var log = logger.NewLogger()

// Output represents the final decoded EXIF data from an image
type Output struct {
	Date      string
	Lng       float64
	Lat       float64
	Copyright string
	FmtData   map[exif.FieldName]int
	AllExif   *exif.Exif
}

func TestExifReader(t *testing.T) {
	t.Run("Run EXIF lat lng", func(t *testing.T) {
		exif.RegisterParsers(mknote.All...)
		for i := 1; i < 4; i++ {
			if i == 0 {
				continue
			}
			path := fmt.Sprintf("./testimgs/%v.jpg", i)
			output, err := GetOutput(path)
			if err != nil {
				t.Error(err)
			}
			spew.Dump(output)
		}
	})
}

func GetOutput(path string) (*Output, error) {
	x := loadExif(path)
	var err error
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

	// fmtData, err := x.getImageFormatData()
	// if err != nil {
	// 	return nil, err
	// }

	return &Output{
		Lat:  lat,
		Lng:  lng,
		Date: date,
		// FmtData: fmtData,
		AllExif: x.x,
	}, nil
}

type exifData struct {
	x *exif.Exif
}

func loadExif(path string) *exifData {
	file, err := os.Open(path)
	if err != nil {
		log.Errorln(err)
	}
	x, err := exif.Decode(file)
	if err != nil {
		log.Errorln(err)
	}
	return &exifData{
		x: x,
	}
}

func (e *exifData) calcGeoCoordinate(fieldName exif.FieldName) (float64, error) {
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

func (e *exifData) getDateTime() (date string, err error) {
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

// type imgFmtData struct {
// 	ImageWidth  map[string]int
// 	ImageLength map[string]int
// 	XResolution map[string]int
// 	YResolution map[string]int
// }

func (e *exifData) getImageFormatData() (map[exif.FieldName]int, error) {
	var fNames = []exif.FieldName{"PixelXDimension", "PixelYDimension"}
	var fmtData = make(map[exif.FieldName]int, len(fNames))

	for _, n := range fNames {
		fName := exif.FieldName(n)
		tag, err := e.x.Get(fName)
		if err != nil {
			return nil, fmt.Errorf("Could not get FieldName tag: %s", err)
		}
		s, err := tag.StringVal()
		if err != nil {
			return nil, fmt.Errorf("Error getting value as string: %s", err)
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("Error converting to integer: %s", err)
		}
		spew.Dump(val)
		fmtData[fName] = val
	}
	return fmtData, nil
}
