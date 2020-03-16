package video

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type video struct{}

func TestEncoder(t *testing.T) {
	t.Run("encode", func(t *testing.T) {
		var v video
		f, _ := os.Open("../in.mp4")

		meta, err := v.RawMetaString(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		spew.Dump(meta)
	})
}

func (v *video) RawMetaString(path string) ([]byte, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("ffmpeg no bin in $PATH")
	}
	// cmd = exec.Command(ffprobe, "-v", "error", "-print_format", "json", "-show_format", "-show_streams", "-hide_banner", r)
	cmd := exec.Command(ffprobe, "-v", "quiet", "-print_format", "json", "-show_format", path)
	return cmd.Output()
}
