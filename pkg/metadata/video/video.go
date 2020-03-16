package video

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type videoMetadata struct {
	io.Reader
}

func New(r io.Reader) *videoMetadata {
	return &videoMetadata{r}
}

func (v *videoMetadata) RawMeta() (*FFMPEGMetaOutput, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("ffprobe no bin in $PATH")
	}
	cmd := exec.Command(ffprobe, "-v", "quiet", "-print_format", "json", "-show_format", "pipe:")
	cmd.Stdin = v.Reader
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

/// ffprobe output
type FFMPEGMetaOutput struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
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
			MajorBrand       string    `json:"major_brand"`
			MinorVersion     string    `json:"minor_version"`
			CompatibleBrands string    `json:"compatible_brands"`
			CreationTime     time.Time `json:"creation_time"`
			Artwork          string    `json:"com.apple.quicktime.artwork"`
			IsMontage        string    `json:"com.apple.quicktime.is-montage"`
			Model            string    `json:"com.apple.quicktime.model"`
			ISOLocation      string    `json:"com.apple.quicktime.location.ISO6709"`
		} `json:"tags"`
	} `json:"format"`
}

func (fo *FFMPEGMetaOutput) SanitizeOutput() *FFMPEGMetaOutput {
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "+", "", 1)
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "+", ",", 1)
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "/", "", -1)
	return fo
}

func (fo *FFMPEGMetaOutput) Height() int {
	return fo.Streams[0].Height
}
func (fo *FFMPEGMetaOutput) Width() int {
	return fo.Streams[0].Width
}
func (fo *FFMPEGMetaOutput) Codec() string {
	return fo.Streams[0].CodecName
}
func (fo *FFMPEGMetaOutput) Lat() string {
	return strings.Split(fo.ISOLocation(), ",")[0]
}
func (fo *FFMPEGMetaOutput) Lng() string {
	return strings.Split(fo.ISOLocation(), ",")[1]
}
func (fo *FFMPEGMetaOutput) Model() string {
	return fo.Format.Tags.Model
}
func (fo *FFMPEGMetaOutput) ISOLocation() string {
	return fo.Format.Tags.ISOLocation
}
func (fo *FFMPEGMetaOutput) CreationTime() time.Time {
	return fo.Format.Tags.CreationTime
}
func (fo *FFMPEGMetaOutput) EndTime() string {
	return fo.Format.Duration
}
func (fo *FFMPEGMetaOutput) StartTime() string {
	return fo.Format.StartTime
}

// ! not in use
func parseLocation(t time.Time) (time.Time, error) {
	tl, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return t, err
	}
	fmt.Println(tl.String())
	return t.In(tl), nil
}
