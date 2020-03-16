package video

import (
	"os"
	"testing"

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

		spew.Dump(meta.Lat())
		spew.Dump(meta.Lng())
	})
}
