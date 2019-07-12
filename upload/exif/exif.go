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

// ImageReader contains image info
type ImageReader struct {
	Image  image.Image
	Name   string
	Format string
	Buffer *bytes.Buffer
}

// ImgService contains methods for imgs
type ImgService interface {
	TagExif(*sync.WaitGroup, chan<- *exif.Exif)
	TagExifSync() *exif.Exif                         // For tests no goroutines
	TagExifError(*sync.WaitGroup, chan<- *TagResult) // Tests
}

var log = logger.NewLogger()

// NewExifReq request exif data for image
func NewExifReq(r io.Reader) (ImgService, error) {
	var buf = new(bytes.Buffer)
	teeRead := io.TeeReader(r, buf)
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
		Image:  src,
		Format: format,
		Name:   string(uuid)[:13] + "." + format,
		Buffer: buf,
	}, nil
}

// TagResult struct for exif channel to return either result or err
type TagResult struct {
	res *exif.Exif
	err error
}

// TagExif returns the bytes of the image/tiff in ch
func (img *ImageReader) TagExif(wg *sync.WaitGroup, ch chan<- *exif.Exif) {
	defer wg.Done()
	out, err := exif.Decode(img.Buffer)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Fatalf("exif.Decode, critical error: %v", err)
		}
		log.Printf("exif.Decode, warning: %v", err)
	}
	log.Printf("Tagged exif: %s", img.Name)
	ch <- out
	close(ch)
}

// TagExifError - testing with ch errors
func (img *ImageReader) TagExifError(wg *sync.WaitGroup, ch chan<- *TagResult) {
	defer wg.Done()
	requiredExifData := [3]exif.FieldName{"DateTime", "GPSLatitude", "GPSLongitude"}
	var tag TagResult
	out, err := exif.Decode(img.Buffer)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Fatalf("exif.Decode, critical error: %v", err)
		}
		log.Printf("exif.Decode, warning: %v", err)
		tag.err = err
	}
	log.Printf("Tagged exif: %s", img.Name)
	for _, rq := range requiredExifData {
		_, err := out.Get(rq)
		if err != nil {
			tag.err = err
		}
	}
	tag.res = out
	ch <- &tag
	close(ch)
}

// TagExifSync returns the bytes of the image/tiff in ch - dont use in production
func (img *ImageReader) TagExifSync() *exif.Exif {
	out, err := exif.Decode(img.Buffer)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Errorf("exif.Decode, critical error: %v", err)
		}
		log.Printf("exif.Decode, warning: %v", err)
	}
	log.Printf("Tagged exif: %s", img.Name)

	return out
}
