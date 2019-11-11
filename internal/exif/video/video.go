package video

import (
	"bytes"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	exif "github.com/blixenkrone/gopro/internal/exif"
	"github.com/blixenkrone/gopro/pkg/file"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var (
	log = logger.NewLogger()
	mut sync.RWMutex
)

const (
	width  = "width"
	height = "height"
)

type videoExifData struct {
	File file.FileGenerator
}

func ReadVideo(r io.Reader) (*videoExifData, error) {
	f, err := file.NewFile(r)
	if err != nil {
		return nil, err
	}

	return &videoExifData{
		File: f,
	}, nil
}

func (v *videoExifData) CreateVideoExifOutput() *exif.Output {
	xErr := &exif.Output{ExifErrors: make(map[string]string)}
	// TODO:
	// thumbnail, err := f.makeThumbnail()
	// if err != nil {
	// 	return nil, err
	// }

	// TODO:
	// _, err := v.execMetadata()
	// if err != nil {
	// 	return nil, err
	// }

	size, err := v.File.FileSize()
	if err != nil {
		err = errors.Cause(err)
		xErr.MissingExif("filesize", err)
	}

	wh, err := v.videoWidthHeight()
	if err != nil {
		xErr.MissingExif("dimension", err)
	}

	return &exif.Output{
		MediaSize:       size,
		PixelXDimension: wh[width],
		PixelYDimension: wh[height],
		ExifErrors:      xErr.ExifErrors,
	}
}

// map represents height and width as string values i.e: hw["width"].
func (v *videoExifData) videoWidthHeight() (wh map[string]int, err error) {
	fileSize, err := v.File.FileSize()
	if err != nil {
		return nil, err
	}
	wh = make(map[string]int)
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("error finding exec path for ffprobe")
	}

	log.Infof("filesize: %v, filename: %s", fileSize, v.File.FileName())
	cmd := exec.Command(ffprobe, "-v", "error", "-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", v.File.FileName())
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "error exec cmd ffprobe width/height")
	}
	regex := `[0-9]*`
	findxRegex := regexp.MustCompile(regex)
	matched, err := regexp.Match(regex, output)
	if err != nil {
		return nil, errors.Wrap(err, "regex failed to match")
	}
	if matched {
		log.Infof("matched %s", v.File.FileName())
		// width always comes first
		dimensions := findxRegex.FindAllString(string(output), -1)
		log.Infof("%s", dimensions)
		arr := []string{width, height}
		for i := range arr {
			if err != nil {
				err = errors.Wrapf(err, "str conv in wh loop failed at pos %v with video val %s", i, dimensions[i])
			}
			dimension, err := strconv.Atoi(dimensions[i])
			if err != nil {
				log.Error(errors.Wrap(err, "dimension strconv failed"))
			}
			wh[arr[i]] = dimension
		}
	} else {
		return nil, errors.New("error finding valid string match for video width/height")
	}
	spew.Dump(wh)
	return wh, err
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
	fsize, err := v.File.FileSize()
	if err != nil {
		return nil, errors.Cause(err)
	}
	log.Infof("size: %d, name: %s", fsize, v.File.FileName())
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
	// spew.Dump(cmd.String())
	// var out bytes.Buffer
	// var stderr bytes.Buffer
	// cmd.Stdout = &out
	// cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "error exec output")
	}
	// spew.Dump(stderr)
	spew.Dump(cmd.Output())
	return string(""), nil
}

// TODO: For later use maybe?
func execCmd(path string, arg ...string) (*exec.Cmd, error) {
	pathcmd, err := exec.LookPath(path)
	if err != nil {
		return nil, errors.Errorf("error finding exec path for %s", path)
	}
	return exec.Command(pathcmd, arg...), nil
}
