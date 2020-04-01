package metadata

import (
	"strings"
)

var (
	VideoFormatSuffix = []string{"mp4", "mov", "quicktime", "x-m4v", "m4v"}
	ImageFormatSuffix = []string{"jpg", "jpeg", "png"}
)

func SupportedVideoSuffix(fileName string) bool {
	fileSuffix := strings.Split(fileName, "/")[1]
	for _, suffix := range VideoFormatSuffix {
		if fileSuffix == suffix {
			return true
		}
	}
	return false
}

func SupportedImageSuffix(fileName string) bool {
	split := strings.Split(fileName, ".")
	if len(split) > 0 {
		fileSuffix := split[len(split)-1]
		for _, suffix := range ImageFormatSuffix {
			if fileSuffix == suffix {
				return true
			}
		}
	}
	return false
}
