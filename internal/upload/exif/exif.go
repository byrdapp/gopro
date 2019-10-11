package exif

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/blixenkrone/gopro/pkg/conversion"
	"github.com/blixenkrone/gopro/pkg/logger"

	_ "image/jpeg"
	_ "image/png"

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
	MediaSize       float64 `json:"mediaSize,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
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
	size, err := x.getFileSize(r)
	if err != nil {
		return nil, fmt.Errorf("Error getting media filesize")
	}

	return &Output{
		Lat:             lat,
		Lng:             lng,
		Date:            date,
		Model:           model,
		PixelXDimension: fmtMap[exif.PixelXDimension],
		PixelYDimension: fmtMap[exif.PixelYDimension],
		Copyright:       author,
		MediaSize:       size,
		// ? do this MediaFormat:     mediaFmt,
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
	d = conversion.UnixNanoToMillis(t)
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

// get file size
func (e *imgExifData) getFileSize(r io.Reader) (float64, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	n, err := buf.Write(b)
	if err != nil {
		return 0, err
	}
	log.Info(n)
	size := conversion.FileSizeBytesToFloat(n)
	log.Info(size)
	return size, nil
}

// get image fmt
// ! switch between image and video - evt create struct input
func (e *imgExifData) getMediaFmt(r io.Reader) (fmt string, err error) {
	// _, fmt, err = image.DecodeConfig(r)
	// if err != nil {
	// 	log.Errorln(err)
	// 	return "", err
	// }
	fmt = ".jpeg"
	return fmt, err
}
