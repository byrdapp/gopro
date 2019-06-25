package exif_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	exifsrv "github.com/blixenkrone/gopro/upload/exif"

	"github.com/rwcarlsen/goexif/exif"
)

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

		imgsrv, err := exifsrv.NewExifReq(r)
		if err != nil {
			t.Errorf("Error getting exif: %s\n", err)
		}

		ch := make(chan *exif.Exif)
		wg.Add(1)
		go func() {
			imgsrv.TagExif(&wg, ch)
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
