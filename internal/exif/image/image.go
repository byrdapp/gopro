package exif

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/blixenkrone/gopro/internal/exif"
	"github.com/blixenkrone/gopro/pkg/conversion"
	"github.com/blixenkrone/gopro/pkg/logger"

	_ "image/jpeg"
	_ "image/png"

	goexif "github.com/rwcarlsen/goexif/exif"
)

var (
	log = logger.NewLogger()
)

// tiff.Tag struct return values as number(i.e. 0 == int)
const (
	exifIntVal = iota
)

type imgExifData struct {
	x *goexif.Exif
}

// GetOutput returns the struct *Output containing img data. Call this for each img.
func ReadImage(r io.Reader) *exif.Output {
	x, err := loadExifData(r)
	if err != nil {
		err = fmt.Errorf("Error loading exif: %s", err)
	}
	lat, err := x.calcGeoCoordinate(goexif.GPSLatitude)
	if err != nil {
		err = fmt.Errorf("Error getting lat data: %s", err)
	}
	lng, err := x.calcGeoCoordinate(goexif.GPSLongitude)
	if err != nil {
		err = fmt.Errorf("Error getting lng data: %s", err)
	}
	date, err := x.getDateTime()
	if err != nil {
		err = fmt.Errorf("Error getting datetime: %s", err)
	}
	author, err := x.getCopyright()
	if err != nil {
		err = fmt.Errorf("Error getting copyright: %s", err)
	}
	model, err := x.getCameraModel()
	if err != nil {
		err = fmt.Errorf("Error getting camera model: %s", err)
	}
	fmtMap, err := x.getImageFormatData()
	if err != nil {
		err = fmt.Errorf("Error getting img fmt data: %s", err)
	}
	size, err := x.getFileSize(r)
	if err != nil {
		err = fmt.Errorf("Error getting media filesize")
	}

	return &exif.Output{
		Lat:             lat,
		Lng:             lng,
		Date:            date,
		Model:           model,
		PixelXDimension: fmtMap[goexif.PixelXDimension],
		PixelYDimension: fmtMap[goexif.PixelYDimension],
		Copyright:       author,
		MediaSize:       size,
		// ? do this MediaFormat:     mediaFmt,
	}
}

// loadExifData request exif data for image
func loadExifData(r io.Reader) (*imgExifData, error) {
	x, err := goexif.Decode(r)
	if err != nil {
		log.Errorln("ERROR DECODING: " + err.Error())
		return nil, fmt.Errorf("Error decoding EXIF in image")
	}
	return &imgExifData{x}, nil
}

func (e *imgExifData) calcGeoCoordinate(fieldName goexif.FieldName) (float64, error) {
	tag, err := e.x.Get(fieldName)
	if err != nil {
		if goexif.IsTagNotPresentError(err) {
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
	tag, err := e.x.Get(goexif.Copyright)
	if err != nil {
		return author, err
	}
	return tag.StringVal()
}

func (e *imgExifData) getCameraModel() (model string, err error) {
	n := goexif.FieldName(goexif.Model)
	tag, err := e.x.Get(n)
	if err != nil {
		return model, err
	}
	return tag.StringVal()
}

func (e *imgExifData) getImageFormatData() (map[goexif.FieldName]int, error) {
	var fNames = []goexif.FieldName{goexif.PixelXDimension, goexif.PixelYDimension}
	var fNameVal = make(map[goexif.FieldName]int, len(fNames))
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
	size := conversion.FileSizeBytesToFloat(n)
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
