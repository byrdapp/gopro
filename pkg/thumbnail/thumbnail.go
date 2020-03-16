package thumbnail

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	fromSecondMark = "00:00:01.000"
	toSecondMark   = "00:00:01.100"
)

type thumbnail struct {
	r io.Reader
}

func New(r io.Reader) *thumbnail {
	return &thumbnail{r}
}

type FFMPEGThumbnail []byte

func (t *thumbnail) Thumbnail(x, y int) (FFMPEGThumbnail, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("ffmpeg no bin in $PATH")
	}
	f, err := ioutil.TempFile(os.TempDir(), "video-*")
	if err != nil {
		return nil, err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	// -v quiet -i ./in.mp4 -ss 00:00:01.000 -vframes 1 -s 300x300 out.jpg
	cmd := exec.Command(ffmpeg, "-i", "pipe:", "-ss", fromSecondMark, "-to", toSecondMark, "-vframes", "1", "-s", fmt.Sprintf("%vx%v", x, y), "-f", "singlejpeg", "pipe:")
	fmt.Println(cmd.String())
	cmd.Stdin = t.r
	return cmd.CombinedOutput()
}

func removeFile(f *os.File) error {
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Remove(f.Name()); err != nil {
		return err
	}
	return nil
}
