package video

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"

	exif "github.com/blixenkrone/gopro/pkg/exif"
	"github.com/blixenkrone/gopro/pkg/file"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var (
	log = logger.NewLogger()
)

const (
	width  = "width"
	height = "height"
	lat    = "lat"
	lng    = "lng"
)

type videoFile struct {
	File *file.File
}

func ReadVideo(r io.Reader) (*videoFile, error) {
	f, err := file.NewFile(r)
	if err != nil {
		return nil, err
	}

	return &videoFile{
		File: f,
	}, nil
}

func (v *videoFile) CreateVideoExifOutput() *exif.Output {
	xErr := &exif.Output{MissingExif: make(map[string]string)}
	// TODO:
	// thumbnail, err := f.makeThumbnail()
	// if err != nil {
	// 	return nil, err
	// }
	size, err := v.File.FileSize()

	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("filesize", err)
	}

	meta, err := v.ffprobeVideoMeta()
	if err != nil {
		xErr.AddMissingExif("metacmd", err)
	}

	// log.Info("FFPROBE META:\n")
	// spew.Dump(meta)

	geo, err := v.parseLocation(meta.Format.Tags.Location)
	if err != nil {
		xErr.AddMissingExif("geo", err)
		log.Errorf("cause: %s", errors.Cause(err))
	}

	return &exif.Output{
		MediaSize:       size,
		Date:            meta.Format.Tags.CreationTime.UnixNano(),
		Lat:             geo[lat],
		Lng:             geo[lng],
		PixelXDimension: meta.Streams[0].Width,
		PixelYDimension: meta.Streams[0].Height,
		MissingExif:     xErr.MissingExif,
	}
}

func (v *videoFile) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, v.File.File())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *videoFile) ExtractImage() ([]byte, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, v.File.File())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type ffprobeOutput struct {
	Format  *ffprobeFormat   `json:"format,omitempty"`
	Streams []*ffprobeStream `json:"streams,omitempty"`
}

type ffprobeStream struct {
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration string `json:"duration,omitempty"`
}

type ffprobeFormat struct {
	Filename string       `json:"filename,omitempty"`
	Duration string       `json:"duration,omitempty"`
	Size     string       `json:"size,omitempty"`
	Tags     *ffprobeTags `json:"tags,omitempty"`
}

type ffprobeTags struct {
	CreationTime time.Time `json:"creation_time,omitempty"`
	Location     string    `json:"location,omitempty"`
}

// func (v *videoExifData) ffprobeVideoMeta() ([]byte, error) {
// 	ffmpeg, _ := exec.LookPath("ffmpeg")

// 	cmd := exec.Command(ffmpeg, "-i", v.File.FileName(), "-c:v", "copy", "-movflags", "faststart", "-strict", "-2", "-moov_size", "bytes", "-hide_banner")
// 	log.Info(cmd.String())
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "error cmd stdout: %s", out)
// 	}
// 	return out, nil
// }

func (v *videoFile) ffprobeVideoMeta() (*ffprobeOutput, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("error finding exec path for ffprobe")
	}

	cmd := exec.Command(ffprobe, "-v", "error", "-print_format", "json", "-show_format", "-show_streams", "-hide_banner", v.File.FileName())
	log.Info(cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "error cmd stdout: %s", out)
	}

	var ffprobeOut ffprobeOutput
	if err := json.Unmarshal(out, &ffprobeOut); err != nil {
		return nil, errors.Cause(err)
	}
	return &ffprobeOut, nil
}

// Parse geolocation from single video file - no array pls
func (v *videoFile) parseLocation(input string) (geo map[string]float64, err error) {
	if len(input) <= 0 {
		return nil, errors.New("error: no location data provided in file")
	}
	geo = make(map[string]float64, 2)
	ISOCoordinate, match := matchRegExp(`((\+|-)\d+\.?\d*)`, []byte(input))
	if !match {
		return nil, errors.New("regex failed to match")
	}
	coordinates := []string{lat, lng}
	result := ISOCoordinate.FindAllString(input, 2)
	for i := range result {
		f, err := strconv.ParseFloat(result[i], 64)
		if err != nil {
			return nil, errors.Errorf("error parsing %s as float", coordinates[i])
		}
		geo[coordinates[i]] = f
	}
	return geo, nil
}

// map represents height and width as string values i.e: hw["width"].
func (v *videoFile) matchWHDimension(input []byte) (wh map[string]int, err error) {
	exp, match := matchRegExp(`[0-9]*`, input)
	if !match {
		return nil, errors.New("regex failed to match")
	}

	log.Infof("matched %s", v.File.FileName())
	wh = make(map[string]int)
	// width always comes first
	dimensions := exp.FindAllString(string(input), -1)
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
	return wh, err
}

func matchRegExp(str string, b []byte) (*regexp.Regexp, bool) {
	exp := regexp.MustCompile(str)
	if matched := exp.Match(b); !matched {
		return nil, false
	} else {
		return exp, true
	}
}
