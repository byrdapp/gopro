package video

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	VideoFormatSuffix = []string{"mp4", "mov", "quicktime", "x-m4v", "m4v", "jpeg"}
	fromSecondMark    = "00:00:01.000"
	toSecondMark      = "00:00:01.100"
)

func SupportedSuffix(fileName string) bool {
	fileSuffix := strings.Split(fileName, "/")[1]
	for _, suffix := range VideoFormatSuffix {
		if fileSuffix == suffix {
			return true
		}
	}
	return false
}

type VideoReader interface {
	io.Reader
}

func RawMeta(r VideoReader) (*FFMPEGMetaOutput, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("ffprobe no bin in $PATH")
	}
	cmd := exec.Command(ffprobe, "-v", "quiet", "-print_format", "json", "-show_format", "pipe:")
	cmd.Stdin = r
	outJSON, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	var ffmpeg FFMPEGMetaOutput
	if err := json.Unmarshal(outJSON, &ffmpeg); err != nil {
		return nil, err
	}
	return &ffmpeg, nil
}

func (fo *FFMPEGMetaOutput) SanitizeOutput() *FFMPEGMetaOutput {
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "+", "", 1)
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "+", ",", 1)
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "/", "", -1)
	return fo
}

// ffprobe output
type FFMPEGMetaOutput struct {
	Format struct {
		Filename       string `json:"filename"`
		NbStreams      int    `json:"nb_streams"`
		NbPrograms     int    `json:"nb_programs"`
		FormatName     string `json:"format_name"`
		FormatLongName string `json:"format_long_name"`
		StartTime      string `json:"start_time"`
		Duration       string `json:"duration"`
		ProbeScore     int    `json:"probe_score"`
		Tags           struct {
			MajorBrand                 string    `json:"major_brand"`
			MinorVersion               string    `json:"minor_version"`
			CompatibleBrands           string    `json:"compatible_brands"`
			CreationTime               time.Time `json:"creation_time"`
			ComAppleQuicktimeArtwork   string    `json:"com.apple.quicktime.artwork"`
			ComAppleQuicktimeIsMontage string    `json:"com.apple.quicktime.is-montage"`
			ISOLocation                string    `json:"com.apple.quicktime.location.ISO6709"`
		} `json:"tags"`
	} `json:"format"`
}

type FFMPEGThumbnail []byte

func Thumbnail(r VideoReader, x, y int) (FFMPEGThumbnail, error) {
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
	cmd.Stdin = r
	return cmd.CombinedOutput()
}

func CollectedOutput(r VideoReader) (interface{}, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("ffmpeg no bin in $PATH")
	}
	cmd1 := exec.Command(ffmpeg, "-v", "error", "-i", "pipe:", "-ss", "00:00:01.00", "-vframes", "1", "-s", "300x300", "pipe:")
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("ffprobe no bin in $PATH")
	}
	cmd2 := exec.Command(ffprobe, "-v", "quiet", "-print_format", "json", "-show_format", "pipe:0")
	cmd1Stdin, _ := cmd1.StdinPipe()
	cmd2Stdin, _ := cmd2.StdinPipe()
	mw := io.MultiWriter(cmd1Stdin, cmd2Stdin)
	io.Copy(mw, r)
	return nil, nil
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

func parseLocation(t time.Time) (time.Time, error) {
	tl, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return t, err
	}
	fmt.Println(tl.String())
	return t.In(tl), nil
}
