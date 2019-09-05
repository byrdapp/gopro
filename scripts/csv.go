package scripts

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

// CreateCSVWriterFile asd
func CreateCSVWriterFile(fileName string) *csv.Writer {
	err := os.Chdir("./csv-files")
	if err != nil {
		log.Println(err)
	}
	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
	}
	w := csv.NewWriter(file)
	return w
}

// WriteColumnHeaders sizes of columns
func WriteColumnHeaders(columns []string, csvWriter *csv.Writer) {
	err := csvWriter.Write(columns)
	if err != nil {
		log.Fatalf("Error writing columns %s", err)
	}
}

// ParseUnixAsDate asd
func ParseUnixAsDate(val int64) string {
	if val == 0 {
		return "%"
	}
	date := time.Unix((val / 1000), 0)
	fmt.Println(date)
	return date.UTC().String()
}

// ParseFloatUnixDate float64 instead of int64
func ParseFloatUnixDate(t *time.Time) string {
	return t.UTC().String()
}
