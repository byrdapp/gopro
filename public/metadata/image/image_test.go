package image

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/byrdapp/byrd-pro-api/internal/storage/aws"
	"github.com/byrdapp/byrd-pro-api/public/metadata"
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
	var output []*metadata.Metadata
	for i := 1; i < 5; i++ {
		fileName := fmt.Sprintf("%v.jpg", i)
		t.Run("success exif", func(t *testing.T) {
			mat, err := aws.GetTestMaterial(aws.ImageBucketReference, fileName)
			if err != nil {
				t.Error(err)
			}
			log.Info("got bytes")
			r := bytes.NewReader(mat.Buf.Bytes())
			_, err = ImageMetadata(r)
			if err != nil {
				if !equalError(t, err, metadata.EOFError.Error()) {
					t.Fatal(err)
				}
			}
			log.Info("parsed exif")
			o := &metadata.Metadata{}

			output = append(output, o)
		})
	}
}

func equalError(t *testing.T, input error, expectedMsg string) bool {
	return assert.EqualError(t, input, expectedMsg)
}
