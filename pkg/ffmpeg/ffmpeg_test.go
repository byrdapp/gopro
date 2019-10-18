package ffmpeg

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"
)

type videoExif struct {
	CreationTime time.Time `tag:"creation_time"`
	Location     location  `tag:"location"`
}

// VideoOutput -
type VideoOutput struct {
	Buffer    bytes.Reader
	Video     os.File    `json:"video"`
	Thumbnail os.File    `json:"thumbnail"`
	VideoExif *videoExif `json:"exif"`
}

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
		var err error
		file := "./video/in.mp4"
		output := "./video/tmp/output"
		ffmpeg, err := exec.LookPath("ffmpeg")
		if err != nil {
			t.Log(err)
		}
		// test cmd: $ go test -v pkg/ffmpeg/ffmpeg_test.go
		cmd := exec.Command(ffmpeg, "-y", "-i", file, "-ss", "00:00:08", "-vframes", "1", output+".png", "-f", "ffmetadata", "-map_metadata", "0", output+".txt")
		t.Log(cmd.String())

		var out bytes.Buffer
		var stderr bytes.Buffer

		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			t.Log(err)
			t.Log(stderr.String())
			return
		}

		t.Log(out.String())
	})
}

type ReaderOutput interface {
	Read() error
	OutputImage()
}

func (v *VideoOutput) Read() error {
	return nil
}

func (v *VideoOutput) OutputImage() {
	// file, err := os.OpenFile(v.Thumbnail, 1, os.ModeAppend)
	// if err != nil {
	// 	return err
	// }

	// bufio.NewScanner(file)
}
