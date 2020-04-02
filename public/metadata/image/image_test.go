package image

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/byrdapp/byrd-pro-api/internal/storage/aws"
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
	fileName := "1.jpg"
	t.Run("success exif", func(t *testing.T) {
		mat, err := aws.GetTestMaterial(aws.ImageBucketReference, fileName)
		if err != nil {
			t.Error(err)
		}
		log.Info("got bytes")
		r := bytes.NewReader(mat.Buf.Bytes())
		meta, err := ImageMetadata(r)
		if err != nil {
			t.Fatal(err)
		}
		spew.Dump(meta.x.String())
	})
}

func equalError(t *testing.T, input error, expectedMsg string) bool {
	return assert.EqualError(t, input, expectedMsg)
}
