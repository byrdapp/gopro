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
	Image        image.Image
	Name         string
	Format       string
	ByteVal      []byte
	Buffer       *bytes.Buffer
	BufferReader *bytes.Reader
}

// ImgService contains methods for imgs
type ImgService interface {
	TagExif(*sync.WaitGroup, chan<- *exif.Exif)
	TagExifSync() *exif.Exif // For tests no goroutines
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

// TagExifSync returns the bytes of the image/tiff in ch - dont use in production
func (img *ImageReader) TagExifSync() *exif.Exif {
	out, err := exif.Decode(img.Buffer)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Fatalf("exif.Decode, critical error: %v", err)
		}
		log.Printf("exif.Decode, warning: %v", err)
	}
	log.Printf("Tagged exif: %s", img.Name)

	return out
}
