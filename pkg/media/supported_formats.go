package media

import (
	"errors"
	"strings"
)

type FileFormat string

var ErrUnsupportedFormat = errors.New("media format is not supported")

const (
	MP4       = "mp4"
	MOV       = "mov"
	Quicktime = "quicktime" // ? this is .mov basically
	JPEG      = "jpeg"
	JPG       = "jpg"
)

func lowercaseFmt(s FileFormat) FileFormat {
	return FileFormat(strings.ToLower(string(s)))
}

func Format(inputFormat string) (_ FileFormat) {
	return FileFormat(inputFormat)
}

func (f FileFormat) Video() (FileFormat, error) {
	if _, ok := videoFormats[lowercaseFmt(f)]; !ok {
		return "", ErrUnsupportedFormat
	} else {
		return f, nil
	}
}

func (f FileFormat) Image() (FileFormat, error) {
	if _, ok := imageFormats[lowercaseFmt(f)]; !ok {
		return "", ErrUnsupportedFormat
	} else {
		return f, nil
	}
}

var videoFormats = map[FileFormat]FileFormat{
	MP4:       "mp4",
	MOV:       "mov",
	Quicktime: "quicktime",
}

var imageFormats = map[FileFormat]FileFormat{
	JPEG: "jpeg",
	JPG:  "jpg",
}
