package exif

import (
	"bytes"
	"fmt"
	"image"

	"github.com/byblix/gopro/utils/logger"

	// Keep this import so the compiler knows the format
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os/exec"
	"sync"

	"github.com/rwcarlsen/goexif/exif"
)

/**
 * ! Only takes JPEG as image format atm
 */

// ImageReader contains image info
type ImageReader struct {
	Image   image.Image
	Name    string
	Format  string
	ByteVal []byte
	Reader  io.Reader
}

// Service contains methods
type Service interface {
	TagExif(*sync.WaitGroup, chan<- []byte)
}

var log = logger.NewLogger()

// NewExif request exif data for image
func NewExif(r io.Reader) (Service, error) {
	var buf bytes.Buffer
	teeRead := io.TeeReader(r, &buf)
	src, format, err := image.Decode(teeRead)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Image format is: %s\n", format)

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		return nil, err
	}

	return &ImageReader{
		Image:   src,
		Format:  format,
		Name:    string(uuid)[:13] + "." + format,
		ByteVal: buf.Bytes(),
		Reader:  &buf,
	}, nil
}

// TagExif returns the bytes of the image/tiff in ch
func (img *ImageReader) TagExif(wg *sync.WaitGroup, ch chan<- []byte) {
	defer wg.Done()
	out, err := exif.Decode(img.Reader)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Fatalf("exif.Decode, critical error: %v", err)
		}
		log.Printf("exif.Decode, warning: %v", err)
	}
	log.Printf("Tagged exif: %s", img.Name)
	val, err := out.MarshalJSON()
	if err != nil {
		log.Fatalf("Error marshalling JSON: %s", err)
	}
	ch <- val
	close(ch)
}
