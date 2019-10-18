package ffmpeg

import (
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/blixenkrone/gopro/pkg/ffmpeg"
)

// VideoOutput -

// const ErrFileOpen int = 0

func TestFFMPEGVideo(t *testing.T) {
	// t.Run("ffmpeg lib", func(t *testing.T) {
	// 	file := "./video/in.mp4"
	// 	avformat.AvRegisterAll()

	// 	ctx := avformat.AvformatAllocContext()
	// 	if err := avformat.AvformatOpenInput(&ctx, file, nil, nil); err != ErrFileOpen {
	// 		t.Error(err)
	// 	}

	// })
}

func TestVideoLoad(t *testing.T) {
	t.Run("ffmpeg bash", func(t *testing.T) {
		// full cmd: $ ffmpeg -i in.mp4 -ss 00:00:08 -vframes 1 out.png -f ffmetadata -map_metadata 0 metadata.txt
		f, err := os.Open("./video/in.mp4")
		if err != nil {
			t.Log(err)
		}

		ffile, err := ffmpeg.NewFile(f)
		if err != nil {
			t.Log(err)
		}

		out, err := ffile.CreateVideoOutput()
		if err != nil {
			t.Log(err)
		}

		spew.Dump(out)

	})
}
