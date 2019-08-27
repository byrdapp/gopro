package exif_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rwcarlsen/goexif/tiff"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

type TagResult struct {
	Exif *exif.Exif
	Date *time.Time
	Lng  float64
	Lat  float64
	Err  string
}

func TestExifReader(t *testing.T) {
	t.Run("Run EXIF lat lng", func(t *testing.T) {
		exif.RegisterParsers(mknote.All...)
		for i := 1; i < 5; i++ {
			if i == 0 {
				continue
			}
			path := fmt.Sprintf("./test2/%v.jpg", i)
			file, err := os.Open(path)
			if err != nil {
				t.Errorf("Error: %s\n", err)
			}
			x, err := exif.Decode(file)
			// x, err := exif.LazyDecode(file)
			if err != nil {
				t.Errorf("Error: %s", err)
			}

			// lat, _ := x.Get(exif.GPSLongitude)
			// deg, _ := lat.Rat(0)
			// floatDeg, _ := deg.Float64()
			// min, _ := lat.Rat(1)
			// floatMin, _ := min.Float64()
			// sec, _ := lat.Rat(2)
			// floatSec, _ := sec.Float64()
			// floatLat := floatDeg + floatMin/60 + floatSec/3600

			// t.Log(floatLat)

			tRes := &TagResult{}
			fieldNames := [2]exif.FieldName{exif.GPSLatitude, exif.GPSLongitude}
			for _, val := range fieldNames {
				tag, err := x.Get(val)
				if err != nil {
					t.Error(err)
				}
				latlng, err := tRes.setCoordinates(tag)
				if err != nil {
					t.Error(err)
				}
				t.Logf("Values: %v %v", tRes.Lat, tRes.Lng)
			}

		}
	})
}

type LatLng struct {
	Lat float64
	Lng float64
}

func (t *TagResult) setCoordinates(tag *tiff.Tag) (*LatLng, error) {
	intRationals := [3]int{0, 1, 2}
	var finalFloats = make([]float64, 3)
	calc := map[string]int{"deg": 0, "min": 1, "sec": 2}
	res := map[string]float64{"1": 0.0, "2": 0.0, "3": 0.0}

	for key, val := range calc {
		ratVal, err := tag.Rat(num)
		if err != nil {
			return nil, err
		}
		calc["deg"] = ratVal.Float64()
		calc["deg"] = ratVal.Float64()
		calc["deg"] = ratVal.Float64()
		f, _ := ratVal.Float64()
		finalFloats = append(finalFloats, f)
	}

	for _, v := range finalFloats {
		fmt.Println(v)
	}

	return &res, nil
}
