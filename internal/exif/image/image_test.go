package image

import (
	"fmt"
	"testing"

	"github.com/blixenkrone/gopro/internal/exif"
	"github.com/blixenkrone/gopro/internal/storage/aws"
)

func TestImageReader(t *testing.T) {
	if err := aws.ParseCredentials(); err != nil {
		t.Fatal(err)
	}
	t.Run("success exif", func(t *testing.T) {
		var output []*exif.Output
		for i := 0; i < 5; i++ {
			mat, err := aws.GetTestMaterial(aws.ImageBucketReference, fmt.Sprintf("%v.jpg", i))
			if err != nil {
				t.Error(err)
				return
			}
			parsedExif := ReadImage(mat.Bytes())
			output = append(output, parsedExif)
		}
	})
	t.Run("fail exif", func(t *testing.T) {
		var output []*exif.Output
		for i := 0; i < 5; i++ {
			mat, err := aws.GetTestMaterial(aws.ImageBucketReference, "fail.jpg")
			if err != nil {
				t.Error(err)
				return
			}
			parsedExif := ReadImage(mat.Bytes())
			output = append(output, parsedExif)
		}
	})
}
