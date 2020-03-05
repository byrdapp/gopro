package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/blixenkrone/byrd/byrd-pro-api/pkg/logger"
)

const (
	// TimeFormat is the std in Dk
	TimeFormat = time.UnixDate
)

var log = logger.NewLogger()

// TimeBuilder -
type TimeBuilder struct {
	DateStart time.Time `json:"dateStart"`
	DateEnd   time.Time `json:"dateEnd"`
}

// TimeParse exported interface to pkg's
type TimeParse interface {
	UnmarshallJSON([]byte) error
	IsZero() bool
}

// NewTime inits a builder that can call time methods
func NewTime(start, end time.Time) *TimeBuilder {
	return &TimeBuilder{start, end}
}

// UnmarshallJSON -
func (t *TimeBuilder) UnmarshallJSON(r io.Reader) error {
	j, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(j, t); err != nil {
		return err
	}
	log.Infof("Struct decoded: %v\n", t)
	return nil
}

// IsZero checks for bad formatting and null errors
func (t *TimeBuilder) IsZero() error {
	var err error
	if t.DateStart.IsZero() || t.DateEnd.IsZero() {
		err = fmt.Errorf("The start and/or enddate is missing or badly formatted")
		return err
	}
	return nil
}
