package conversion

import (
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// JoinStrings join strings
func JoinStrings(list []string) string {
	return strings.Join(list, ", ")
}

// ParseBool parses string to bool val
func ParseBool(stringValToBool string) (bool, error) {
	return strconv.ParseBool(stringValToBool)
}

const (
	B  = 1
	KB = B << 10
	MB = KB << 10
)

func CalculateFileSize(byteSize int) (ui uint64, err error) {
	str := strconv.Itoa(byteSize)
	ui, err = humanize.ParseBytes(str)
	return ui, err
}

func UnixNanoToMillis(t time.Time) int64 {
	return t.UTC().UnixNano() / int64(time.Millisecond)
}
