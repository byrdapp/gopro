package metadata

import (
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestLatLng(t *testing.T) {
	t.Run("metadata", func(t *testing.T) {
		f, _ := os.Open("./video/media/in.mp4")
		m, err := VideoMetadata(f)
		if err != nil {
			t.Fatal(err)
		}
		spew.Dump(m)
	})
}

func TestMetadata(t *testing.T) {
	t.Run("metadata", func(t *testing.T) {
		f, _ := os.Open("./video/media/in.mp4")
		meta, err := VideoMetadata(f)
		if err != nil {
			t.Fatal(err)
		}
		spew.Dump(meta)
	})
}
