package video

import (
	"os"
	"testing"

	"github.com/byrdapp/byrd-pro-api/pkg/conversion"
	"github.com/davecgh/go-spew/spew"
)

func TestEncoder(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		f, _ := os.Open("./media/in.mp4")
		m := New(f)
		meta, err := m.RawMeta()
		if err != nil {
			t.Fatal(err)
		}
		meta.SanitizeOutput()
		spew.Dump(meta)
	})
}
func TestLatLngConversion(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		f, _ := os.Open("./media/in.mp4")
		m := New(f)
		meta, err := m.RawMeta()
		if err != nil {
			t.Fatal(err)
		}
		meta.SanitizeOutput()
		spew.Dump(conversion.MustStringToFloat(meta.Lat()))
		spew.Dump(conversion.MustStringToFloat(meta.Lng()))
	})
}
func TestWidthHeight(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		f, _ := os.Open("./media/in.mp4")
		m := New(f)
		meta, err := m.RawMeta()
		if err != nil {
			t.Fatal(err)
		}
		meta.SanitizeOutput()
		spew.Dump(meta.Height())
		spew.Dump(meta.Width())
	})
}
