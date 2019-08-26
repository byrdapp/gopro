package exif_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	exifsrv "github.com/blixenkrone/gopro/upload/exif"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

func TestExifReader(t *testing.T) {

	exif.RegisterParsers(mknote.All...)
	file, err := os.Open("./tests/1.jpeg")
	if err != nil {
		t.Errorf("Error: %s\n", err)
	}

	x, err := exif.LazyDecode(file)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	lat, lng, err := x.LatLong()
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	t.Log(lat)
	t.Log(lng)
	// model, err := x.Get(exif.Model)
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Logf("Model %s", model.String())

	// lat, err := x.Get(exif.GPSLatitude)
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Logf("Lat: %s", lat)

	// long, err := x.Get(exif.GPSLongitude)
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Logf("Long: %s", long)

	// lat, lng, err := x.LatLong()
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Logf("Lat: %v, Lng: %v", lat, lng)

	t.Log(x.String())
}

var wg sync.WaitGroup

func TestMultipleImageReader(t *testing.T) {
	var exifs []*exif.Exif
	for i := 0; i < 5; i++ {
		if i == 0 {
			continue
		}
		fmt.Printf("Running image: %d", i)
		fmtFileName := fmt.Sprintf("../../assets/teststories/image%v.jpeg", i)
		fileName, err := filepath.Abs(fmtFileName)
		if err != nil {
			t.Errorf("Couldnt read error: %s", err)
		}
		fbytes, err := ioutil.ReadFile(fileName)
		if err != nil {
			t.Errorf("Error: %s\n", err)
		}

		r := bytes.NewReader(fbytes)

		_, err = exifsrv.NewExifReq(r)
		if err != nil {
			t.Errorf("Error getting exif: %s\n", err)
		}

		ch := make(chan *exif.Exif)
		wg.Add(1)
		go func() {
			// imgsrv.TagExif(&wg, ch)
		}()
		exif := <-ch
		exifs = append(exifs, exif)
	}
	wg.Wait()
	for idx, val := range exifs {
		t.Logf("Index: %v, value: %s\n\n", idx, val)
	}
	// rr := httptest.NewRecorder()
	// byt, _ := json.Marshal(exifs)
	// r := bytes.NewReader(byt)
	// req := httptest.NewRequest("POST", "/exif", r)
	// body, _ := ioutil.ReadAll(req.Body)
	// rr.Write(body)
	t.Log("DONE")
}
