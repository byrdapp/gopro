package ffmpeg

import (
	"os/exec"
	"testing"
)

// const ErrFileOpen int = 0

// func TestFFMPEGVideo(t *testing.T) {
// 	t.Run("ffmpeg lib", func(t *testing.T) {
// 		file := "./video/in.mp4"
// 		avformat.AvRegisterAll()

// 		ctx := avformat.AvformatAllocContext()
// 		if err := avformat.AvformatOpenInput(&ctx, file, nil, nil); err != ErrFileOpen {
// 			t.Error(err)
// 		}

// 	})
// }

func TestVideoLoad(t *testing.T) {
	t.Run("ffmpeg bash", func(t *testing.T) {
		// full cmd: ffmpeg -i in.mp4 -ss 00:00:08 -vframes 1 out.png -f ffmetadata -map_metadata 0 metadata.txt
		var err error
		// var tt time.Time
		file := "./pkg/ffmpeg/video/in.mp4"
		output := "./pkg/ffmpeg/video/tmp/output"
		cmd := exec.Command("ffmpeg", "-i", file, "-ss", "00:00:08", "-vframes", "1", output+".png", "-f", "ffmetadata", "-map_metadata", "0", output+".txt")
		t.Log(cmd.String())
		err = cmd.Run()
		if err != nil {
			t.Log(err)
			t.Errorf("Error with cmd: %s", err)
		}
	})
}
