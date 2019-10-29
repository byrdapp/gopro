package ffmpeg

import (
	"golang.org/x/tools/present"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/blixenkrone/gopro/pkg/conversion"
	"github.com/blixenkrone/gopro/pkg/logger"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

var log = logger.NewLogger()

type VideoOutput struct {
	Thumbnail       []byte  `json:"thumbnail,omitempty"`
	MediaSize       float64 `json:"mediaSize"`
	PixelXDimension int     `json:"pixelXDimension,omitempty"`
	PixelYDimension int     `json:"pixelYDimension,omitempty"`
	// exif.Output
}

type File struct {
	file *os.File
}

type FileGenerator interface {
	CreateVideoOutput() (*VideoOutput, error)
	Close() error
	RemoveFile() error
	FileName() string
	FileSize() (size float64, err error)
}

func NewFile(r io.Reader) (FileGenerator, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	file, err := ioutil.TempFile(os.TempDir(), "prefix-*")
	if err != nil {
		return nil, errors.Wrap(err, "error creating tmp file")
	}
	if _, err = file.Write(b); err != nil {
		return nil, errors.Wrap(err, "error writing to tmp file")
	}
	return &File{file}, nil
}

func (f *File) CreateVideoOutput() (*VideoOutput, error) {
	// thumbnail, err := f.makeThumbnail()
	// if err != nil {
	// 	return nil, err
	// }

	// meta, err := f.execMetadata()
	// if err != nil {
	// 	return nil, err
	// }

	size, err := f.FileSize()
	if err != nil {
		return nil, err
	}

	return &VideoOutput{

		// Metadata: meta,
		// Thumbnail: thumbnail,
		MediaSize: size,
	}, nil

}

func (f *File) execThumbnail() (thumbnail []byte, err error) {
	// full cmd: $ ffmpeg -i in.mp4 -ss 00:00:08 -vframes 1 out.png -f ffmetadata -map_metadata 0 metadata.txt
	output := "./video/tmp/output"
	// tmpDir := ioutil.TempDir("", "")
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("error finding exec path")
	}
	finfo, err := f.file.Stat()
	if err != nil {
		return nil, err
	}
	log.Info(f.file.Name())
	log.Info(finfo.Size())
	fileName := f.file.Name()
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

func (f *File) execMetadata() (string, error) {
	output := "./video/tmp/output"
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", errors.New("error finding exec path")
	}
	cmd := exec.Command(ffmpeg, "-report", "-y", "-i", f.FileName(), "-f", "ffmetadata", "-map_metadata", "0", output+".txt")
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

func (f *File) makeThumbnail() (thumb []byte, err error) {
	thumb, err = f.execThumbnail()
	if err != nil {
		return nil, err
	}
	return thumb, err
}

func (f *File) videoHeightWidth() {
	present.

}

func (f *File) Close() error {
	return f.file.Close()
}

func (f *File) RemoveFile() error {
	return os.Remove(f.file.Name())
}

func (f *File) FileName() string {
	return f.file.Name()
}

func (f *File) FileSize() (size float64, err error) {
	fInfo, err := f.file.Stat()
	if err != nil {
		return size, err
	}
	size = conversion.FileSizeBytesToFloat(int(fInfo.Size()))
	return size, err
}
