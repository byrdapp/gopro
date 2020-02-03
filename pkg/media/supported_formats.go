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

func lowercaseFmt(s string) FileFormat {
	return FileFormat(strings.ToLower(s))
}

func IsSupportedMediaFmt(inputFormat string) (format string, ok bool, err error) {
	if format, ok = supportedMediaFmt[lowercaseFmt(inputFormat)]; !ok {
		return "", false, ErrUnsupportedFormat
	} else {
		return format, ok, nil
	}
}

var supportedMediaFmt = map[FileFormat]string{
	MP4:       "mp4",
	MOV:       "mov",
	Quicktime: "quicktime",
	JPEG:      "jpeg",
	JPG:       "jpg",
}
