package image

import (
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sync"

	"github.com/byrdapp/byrd-pro-api/public/conversion"
	"github.com/byrdapp/byrd-pro-api/public/logger"

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
	x     *goexif.Exif
	rwmut *sync.RWMutex
}

// loadExifData request exif data for image
func ImageMetadata(r io.Reader) (*imgExifData, error) {
	x, err := goexif.Decode(r)
	var mut sync.RWMutex
	if err != nil {
		return nil, err
	}
	return &imgExifData{x, &mut}, nil
}

func (e *imgExifData) Lock() {
	e.rwmut.Lock()
}
func (e *imgExifData) Unlock() {
	e.rwmut.Unlock()
}

func (e *imgExifData) Geo() (lat float64, lng float64, err error) {
	var out []float64
	gpsarr := []goexif.FieldName{goexif.GPSLatitude, goexif.GPSLongitude}
	ratValues := map[string]int{"deg": 0, "min": 1, "sec": 2}
	fValues := make(map[string]float64, len(ratValues))

	for _, GPSref := range gpsarr {
		geoFieldName := goexif.FieldName(GPSref)
		tag, err := e.x.Get(geoFieldName)
		if err != nil {
			return 0, 0, err
		}
		for key, val := range ratValues {
			v, err := tag.Rat(val)
			if err != nil {
				return 0, 0, err
			}
			f, _ := v.Float64()
			e.Lock()
			fValues[key] = f
			e.Unlock()
		}
		res := fValues["deg"] + (fValues["min"] / 60) + (fValues["sec"] / 3600)
		out = append(out, res)
	}
	return out[0], out[1], nil
}

func (e *imgExifData) DateMillisUnix() (d int64, err error) {
	t, err := e.x.DateTime()
	if err != nil {
		return d, err
	}
	d = conversion.UnixNanoToMillis(t)
	return d, nil
}

func (e *imgExifData) Copyright() (author string, err error) {
	tag, err := e.x.Get(goexif.Copyright)
	if err != nil {
		return author, err
	}
	return tag.StringVal()
}

func (e *imgExifData) Model() (model string, err error) {
	n := goexif.FieldName(goexif.Model)
	tag, err := e.x.Get(n)
	if err != nil {
		return model, err
	}
	return tag.StringVal()
}

func (e *imgExifData) Dimensions() (width int, height int, err error) {
	var fNames = []goexif.FieldName{goexif.PixelXDimension, goexif.PixelYDimension}
	var dim []int
	for _, n := range fNames {
		tag, err := e.x.Get(n)
		if err != nil {
			return 0, 0, err
		}
		i, err := tag.Int(exifIntVal)
		if err != nil {
			return 0, 0, err
		}
		dim = append(dim, i)
	}
	return dim[0], dim[1], nil
}

// get file size
// func (e *imgExifData) getFileSize(r io.Reader) (float64, error) {
// 	b, err := ioutil.ReadAll(r)
// 	if err != nil {
// 		return 0, err
// 	}
// 	var buf bytes.Buffer
// 	n, err := buf.Write(b)
// 	if err != nil {
// 		return 0, err
// 	}
// 	size := conversion.FileSizeBytesToFloat(n)
// 	return size, nil
// }
