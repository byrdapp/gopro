package conversion

import (
	"math"
	"strconv"
	"strings"
	"time"
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
	B       = 1
	KB      = B >> 10
	MB      = KB >> 10
	byteVal = 1024
)

func FileSizeBytesToFloat(byteSize int) float64 {
	size := float64(byteSize) / (math.Pow(byteVal, 2))
	return math.Floor(size*100) / 100 // no magic number
}

func MustStringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func UnixNanoToMillis(t time.Time) int64 {
	return t.UTC().UnixNano() / int64(time.Millisecond)
}
