package image

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blixenkrone/gopro/internal/storage/aws"
	"github.com/blixenkrone/gopro/pkg/media"
)

func TestImageReaderFailed(t *testing.T) {
	if err := aws.ParseCredentials(); err != nil {
		t.Fatal(err)
	}
	// var output []*exif.Output

}

func TestImageReaderSuccess(t *testing.T) {
	if err := aws.ParseCredentials(); err != nil {
		t.Fatal(err)
	}
	var output []*media.Output
	for i := 1; i < 5; i++ {
		fileName := fmt.Sprintf("%v.jpg", i)
		t.Run("success exif", func(t *testing.T) {
			mat, err := aws.GetTestMaterial(aws.ImageBucketReference, fileName)
			if err != nil {
				t.Error(err)
			}
			log.Info("got bytes")
			parsedExif, err := DecodeImageMetadata(mat.Bytes())
			if err != nil {
				if !equalError(t, err, EOFError) {
					t.Fatal(err)
				}
			}
			log.Info("parsed exif")
			output = append(output, parsedExif)
		})
	}
	for i := 0; i < 5; i++ {
		t.Run("fail exif for fail.jpg", func(t *testing.T) {
			mat, err := aws.GetTestMaterial(aws.ImageBucketReference, "fail.jpg")
			if err != nil {
				if !equalError(t, err, EOFError) {
					t.Fatal(err)
				}
			}
			parsedExif, err := DecodeImageMetadata(mat.Bytes())
			if err != nil {
				if !equalError(t, err, EOFError) {
					t.Fatal(err)
				}
			}
			output = append(output, parsedExif)
		})
	}
}

func equalError(t *testing.T, input error, expectedMsg string) bool {
	return assert.EqualError(t, input, expectedMsg)
}
