package video

import (
	"bytes"
	"io"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	exif "github.com/blixenkrone/gopro/internal/exif"
	"github.com/blixenkrone/gopro/pkg/fileinfo"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var log = logger.NewLogger()

const (
	width  = "width"
	height = "height"
)

type VideoDimension int

type videoExifData struct {
	File fileinfo.FileGenerator
}

func ReadVideo(r io.Reader) (*videoExifData, error) {
	f, err := fileinfo.NewFile(r)
	if err != nil {
		return nil, err
	}

	return &videoExifData{
		File: f,
	}, nil
}

func (v *videoExifData) CreateVideoExifOutput() (*exif.Output, error) {
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

	size, err := v.File.FileSize()
	if err != nil {
		return nil, err
	}

	wh, err := v.videoWidthHeight()
	if err != nil {
		return nil, err
	}

	return &exif.Output{
		MediaSize:       size,
		PixelXDimension: int(wh[width]),
		PixelYDimension: int(wh[height]),
	}, nil
}

// map represents height and width as string values i.e: hw["width"].
func (v *videoExifData) videoWidthHeight() (wh map[string]VideoDimension, err error) {
	wh = make(map[string]VideoDimension)
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("error finding exec path for ffprobe")
	}
	cmd := exec.Command(ffprobe, "-v", "error", "-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", v.File.FileName())
	b, err := cmd.Output()
	regex := `[0-9]*`
	findxRegex := regexp.MustCompile(regex)
	matched, err := regexp.Match(regex, b)
	if err != nil {
		return nil, errors.Wrap(err, "regex failed to match")
	}
	if matched {
		arr := []string{width, height}
		str := findxRegex.FindAllString(string(b), -1)
		for i := range arr {
			intvals, err := strconv.Atoi(str[i])
			if err != nil {
				err = errors.Wrapf(err, "strconv in wh loop failed at pos %v with video val %s", i, str[i])
			}
			wh[arr[i]] = VideoDimension(intvals)
		}
	} else {
		return nil, errors.New("error finding valid string match for video width/height")
	}
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

// TODO: For later use maybe?
func execCmd(path string, arg ...string) (*exec.Cmd, error) {
	pathcmd, err := exec.LookPath(path)
	if err != nil {
		return nil, errors.Errorf("error finding exec path for %s", path)
	}
	return exec.Command(pathcmd, arg...), nil
}
