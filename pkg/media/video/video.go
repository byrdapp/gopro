package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/byrdapp/byrd-pro-api/pkg/file"
	"github.com/byrdapp/byrd-pro-api/pkg/image/thumbnail"
	"github.com/byrdapp/byrd-pro-api/pkg/logger"
	media "github.com/byrdapp/byrd-pro-api/pkg/media"
)

var (
	log = logger.NewLogger()
)

const (
	lat            = "lat"
	lng            = "lng"
	fromSecondMark = "00:00:00.000"
	toSecondMark   = "00:00:00.100"
)

// TODO: Add some buffered input to files other than mp4
type VideoBuffer struct {
	bufRd  bytes.Buffer
	file   *file.File
	format media.FileFormat
}

func ReadVideoBuffer(r io.Reader, videoFmt media.FileFormat) (*VideoBuffer, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return nil, errors.Wrap(err, "copy failed")
	}
	f, err := file.NewFile(&buf)
	if err != nil {
		return nil, errors.Wrap(err, "tmp file creation")
	}

	return &VideoBuffer{
		bufRd:  buf,
		file:   f,
		format: videoFmt,
	}, nil
}

func (v *VideoBuffer) Metadata() *media.Metadata {
	xErr := &media.Metadata{MissingExif: make(map[string]string)}
	meta, err := v.ffprobeVideoMeta()
	if err != nil {
		xErr.AddMissingExif("metacmd", err)
	}

	geo, err := v.parseLocation(meta.Format.Tags.Location)
	if err != nil {
		xErr.AddMissingExif("geo", err)
		log.Errorf("cause: %s", errors.Cause(err))
	}

	geoISO, err := v.parseLocation(meta.Format.Tags.ISO6709)
	if err != nil {
		xErr.AddMissingExif("geoISO", err)
		log.Errorf("cause: %s", errors.Cause(err))
	}

	date, err := meta.UnixNano(meta.Format.Tags.CreationTime)
	if err != nil {
		xErr.AddMissingExif("date", err)
	}

	return &media.Metadata{
		// MediaSize:       size,
		Date:            date,
		Lat:             geo[lat],
		Lng:             geo[lng],
		ISOLat:          geoISO[lat],
		ISOLng:          geoISO[lng],
		PixelXDimension: meta.Streams[0].Width,
		PixelYDimension: meta.Streams[0].Height,
		MissingExif:     xErr.MissingExif,
	}
}

func (v *VideoBuffer) Thumbnail() (*thumbnail.ParsedImage, error) {
	b, err := v.ffmpegThumbnail(thumbnail.DefaultWidth, thumbnail.DefaultHeight)
	if err != nil {
		return nil, err
	}
	thumb, err := thumbnail.New(b)
	if err != nil {
		return nil, err
	}
	return thumb.EncodeThumbnail()
}

type ffmpegOutput struct {
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
	ISO6709      string    `json:"com.apple.quicktime.location.ISO6709,omitempty"`
}

func (v *VideoBuffer) ffprobeVideoMeta() (*ffmpegOutput, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("error finding exec path for ffprobe")
	}
	var cmd *exec.Cmd
	switch v.format {
	case media.MOV:
		cmd = exec.Command(ffprobe, "-v", "error", "-print_format", "json", "-show_format", "-show_streams", "-hide_banner", v.file.FileName())
	case media.MP4:
		cmd = exec.Command(ffprobe, "-v", "error", "-print_format", "json", "-show_format", "-show_streams", "-hide_banner", v.file.FileName())
	default:
		return nil, media.ErrUnsupportedFormat
	}

	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "error cmd stdout: %s", b)
	}

	var ffprobeOut ffmpegOutput
	if err := json.Unmarshal(b, &ffprobeOut); err != nil {
		return nil, errors.Cause(err)
	}
	return &ffprobeOut, nil
}

func (v *VideoBuffer) ffmpegThumbnail(x, y int) ([]byte, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("error finding exec path for ffprobe")
	}
	var out bytes.Buffer
	var buferr bytes.Buffer
	// pr, pw := io.Pipe()

	var cmd *exec.Cmd
	switch v.format {
	case media.MP4:
		// ? ffmpeg -f mp4 -i filename.mp4 -ss 00:00:01.000 -to 00:00:02.000 -vframes 1 -s 320x240 -f singlejpeg pipe: | cat > out.jpg
		cmd = exec.Command(ffmpeg, "-f", "mp4", "-i", v.file.FileName(), "-ss", fromSecondMark, "-to", toSecondMark, "-vframes", "1", "-s", fmt.Sprintf("%vx%v", x, y), "-f", "singlejpeg", "pipe:1")
	case media.MOV:
		// f, _ := file.NewEmptyFile()
		// ? cat mov.MOV | ffmpeg -f mov -i pipe:0 -ss 00:00:01.000 -to 00:00:02.000 -vframes 1 -s 320x240 -f singlejpeg out.jpg
		log.Info("MOV FILE")
		cmd = exec.Command(ffmpeg, "-f", "mov", "-i", "pipe:0", "-ss", fromSecondMark, "-to", toSecondMark, "-vframes", "1", "-s", fmt.Sprintf("%vx%v", x, y), "-f", "singlejpeg", "out.jpg")
		log.Info(v.bufRd.Bytes()[:4])
		// cmd.Stdin = &v.bufRd
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Error(err)
		}

		if _, err := io.Copy(stdin, &v.bufRd); err != nil {
			log.Error(err)
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = &buferr

	if err := cmd.Start(); err != nil {
		log.Error(buferr.String())
		log.Error(err)
		return nil, err
	}
	go func() {
		if _, err := io.Copy(&out, stdout); err != nil {
			panic(err)
		}
		log.Info("goroutine copied")
	}()

	if err := cmd.Wait(); err != nil {
		log.Error(err)
		log.Error(buferr.String())
		return nil, err
	}
	log.Info("waiting cmd done")

	return out.Bytes(), nil
}

// Parse geolocation from single video file - no array pls
func (v *VideoBuffer) parseLocation(input string) (geo map[string]float64, err error) {
	if len(input) <= 0 {
		return nil, errors.New("error: no location data provided in file")
	}
	geo = make(map[string]float64, 2)
	ISOCoordinate, match := matchRegExp(`((\+|-)\d+\.?\d*)`, []byte(input))
	if !match {
		return nil, errors.New("regex failed to match")
	}
	coordinates := [2]string{lat, lng}
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

type TimeParser interface {
	UnixNano() int64
	IsZero() bool
}

const unixZeroVal = "time value for image is zero"

func (ff *ffmpegOutput) UnixNano(t TimeParser) (int64, error) {
	if t.IsZero() {
		return 0, errors.New(unixZeroVal)
	}
	return t.UnixNano(), nil
}

func matchRegExp(str string, b []byte) (*regexp.Regexp, bool) {
	exp := regexp.MustCompile(str)
	if matched := exp.Match(b); !matched {
		return nil, false
	} else {
		return exp, true
	}
}

type FileService interface {
	Close() error
	RemoveFile() error
}

func (v *VideoBuffer) Close(f FileService) error {
	return f.Close()
}
func (v *VideoBuffer) RemoveFile(f FileService) error {
	return f.RemoveFile()
}
