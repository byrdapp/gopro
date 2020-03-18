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

	cmd := exec.Command(ffprobe, "-v", "quiet", "-print_format", "json", "-show_format", "-show_entries", "stream=height,width,codec_name,size", "pipe:")
	fmt.Println(cmd.String())
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
		CodecName string `json:"codec_name,omitempty"`
		Width     int    `json:"width,omitempty"`
		Height    int    `json:"height,omitempty"`
	} `json:"streams,omitempty"`
	Format struct {
		Filename       string `json:"filename,omitempty"`
		NbStreams      int    `json:"nb_streams,omitempty"`
		NbPrograms     int    `json:"nb_programs,omitempty"`
		FormatName     string `json:"format_name,omitempty"`
		FormatLongName string `json:"format_long_name,omitempty"`
		StartTime      string `json:"start_time,omitempty"`
		Duration       string `json:"duration,omitempty"`
		ProbeScore     int    `json:"probe_score,omitempty"`
		Tags           struct {
			MajorBrand       string    `json:"major_brand,omitempty"`
			MinorVersion     string    `json:"minor_version,omitempty"`
			CompatibleBrands string    `json:"compatible_brands,omitempty"`
			CreationTime     time.Time `json:"creation_time,omitempty"`
			Artwork          string    `json:"com.apple.quicktime.artwork,omitempty"`
			IsMontage        string    `json:"com.apple.quicktime.is-montage,omitempty"`
			Model            string    `json:"com.apple.quicktime.model,omitempty"`
			ISOLocation      string    `json:"com.apple.quicktime.location.ISO6709,omitempty"`
		} `json:"tags,omitempty"`
	} `json:"format,omitempty"`
}

func (fo *FFMPEGMetaOutput) SanitizeOutput() *FFMPEGMetaOutput {
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "+", "", 1)
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "+", ",", 1)
	fo.Format.Tags.ISOLocation = strings.Replace(fo.Format.Tags.ISOLocation, "/", "", -1)
	return fo
}

func (fo *FFMPEGMetaOutput) Height() int {
	if len(fo.Streams) > 0 {
		return fo.Streams[0].Height
	} else {
		return 0
	}
}

func (fo *FFMPEGMetaOutput) Width() int {
	if len(fo.Streams) > 0 {
		return fo.Streams[0].Width
	}
	return 0
}

func (fo *FFMPEGMetaOutput) Codec() string {
	return fo.Streams[0].CodecName
}

func (fo *FFMPEGMetaOutput) Lat() string {
	if fo.ISOLocation() != "" {
		return strings.Split(fo.ISOLocation(), ",")[0]
	}
	return fo.ISOLocation()
}

func (fo *FFMPEGMetaOutput) Lng() string {
	if fo.ISOLocation() != "" {
		return strings.Split(fo.ISOLocation(), ",")[1]
	}
	return fo.ISOLocation()
}

func (fo *FFMPEGMetaOutput) Model() (string, error) {
	if fo.Format.Tags.Model != "" {
		return fo.Format.Tags.Model, nil
	}
	return "", errors.New("missing model from output")
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
