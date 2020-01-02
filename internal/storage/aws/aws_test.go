package aws

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/blixenkrone/gopro/pkg/file"
)

func TestDownloadImages(t *testing.T) {

	_, err := file.SetEnvFileVars("../../../")
	if err != nil {
		t.Error(err)
	}

	i, err := GetTestMaterial(ImageTestReference, "2.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	spew.Dump(i.fileName)
	t.Log("success")
}
