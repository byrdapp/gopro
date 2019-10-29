package video

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	exif "github.com/blixenkrone/gopro/internal/exif"
	"github.com/blixenkrone/gopro/pkg/fileinfo"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var log = logger.NewLogger()

// type Output struct {
// 	exif *exif.Output
// }

type videoExifData struct {
	File fileinfo.FileGenerator
	x    *exif.Output
}

func ReadVideo(r io.Reader) (*videoExifData, error) {
	f, err := fileinfo.NewFile(r)
	if err != nil {
		return nil, err
	}

	return &videoExifData{
		// x: &exif.Output{},
		File: f,
	}, nil
}

func (o *videoExifData) CreateVideoExifOutput() (*exif.Output, error) {
	// TODO:
	// thumbnail, err := f.makeThumbnail()
	// if err != nil {
	// 	return nil, err
	// }

	// TODO:
	// meta, err := f.execMetadata()
	// if err != nil {
	// 	return nil, err
	// }

	size, err := o.File.FileSize()
	if err != nil {
		return nil, err
	}

	return &exif.Output{
		MediaSize: size,
	}, nil
}

func (v *videoExifData) videoHeightWidth() (interface{}, error) {
	return nil, nil
}

func (v *videoExifData) makeThumbnail() (thumb []byte, err error) {
	thumb, err = v.execThumbnail()
	if err != nil {
		return nil, err
	}
	return thumb, err
}

func (v *videoExifData) execThumbnail() (thumbnail []byte, err error) {
	// full cmd: $ ffmpeg -i in.mp4 -ss 00:00:08 -vframes 1 out.png -f ffmetadata -map_metadata 0 metadata.txt
	output := "./video/tmp/output"
	// tmpDir := ioutil.TempDir("", "")
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("error finding exec path")
	}
	finfo, err := v.File.FileStat()
	if err != nil {
		return nil, err
	}
	log.Info(v.File.FileName())
	log.Info(finfo.Size())
	fileName := v.File.FileName()
	// test cmd: $ go test -v pkg/ffmpeg/ffmpeg_test.go
	cmd := exec.Command(ffmpeg, "-report", "-y", "-i", fileName, "-ss", "00:00:04", "-vframes", "1", output+".png")
	log.Info(cmd.String())
	// cmd2 := exec.Command(ffmpeg, "-f", "ffmetadata", "-map_metadata", "0", output+".txt")
	// cmd := exec.Command(ffmpeg...args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "error running ffmpeg exec cmd: "+stderr.String())
	}
	log.Info(out.Bytes())
	return stderr.Bytes(), nil
}

func (v *videoExifData) execMetadata() (string, error) {
	output := "./video/tmp/output"
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", errors.New("error finding exec path")
	}
	cmd := exec.Command(ffmpeg, "-report", "-y", "-i", v.File.FileName(), "-f", "ffmetadata", "-map_metadata", "0", output+".txt")
	spew.Dump(cmd.String())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "error exec output")
	}
	spew.Dump(stderr)
	return string(""), nil

}
