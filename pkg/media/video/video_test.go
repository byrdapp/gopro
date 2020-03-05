package video

import (
	"bytes"
	"testing"

	"github.com/blixenkrone/byrd/byrd-pro-api/internal/storage/aws"
	"github.com/blixenkrone/byrd/byrd-pro-api/pkg/image/thumbnail"
	"github.com/davecgh/go-spew/spew"
	// "github.com/blixenkrone/byrd/byrd-pro-api/pkg/exif/video"
)

func TestVideoExifBuffer(t *testing.T) {
	if err := aws.ParseCredentials(); err != nil {
		t.Error(err)
		return
	}

	m, err := aws.GetTestMaterial("videos", "mov_small.mov")
	if err != nil {
		t.Error(err)
		return
	}
	rd := bytes.NewReader(m.Bytes())
	video, err := ReadVideoBuffer(rd, "mov")
	if err != nil {
		t.Error(err)
		panic(err)
	}

	thumb, err := video.ffmpegThumbnail(thumbnail.DefaultWidth, thumbnail.DefaultHeight)
	if err != nil {
		t.Error(err)
		return
	}
	img, err := thumbnail.New(thumb)
	if err != nil {
		log.Error(err)
		return
	}
	pthumb, err := img.EncodeThumbnail()
	if err != nil {
		log.Error(err)
		return
	}

	spew.Dump(pthumb.Bytes()[1:2])

	defer func() {
		if err := video.Close(video.file); err != nil {
			log.Error(err)
		}
		if err := video.RemoveFile(video.file); err != nil {
			log.Error(err)
		}
	}()
}

func TestVideoExifTmpFile(t *testing.T) {
	if err := aws.ParseCredentials(); err != nil {
		t.Error(err)
		return
	}
	m, err := aws.GetTestMaterial("videos", "mov.mov")
	if err != nil {
		t.Error(err)
		return
	}
	r := bytes.NewReader(m.Bytes())
	v, err := ReadVideoBuffer(r, "mov")
	if err != nil {
		t.Error(err)
		panic(err)
	}
	b, err := v.ffmpegThumbnail(400, 200)
	if err != nil {
		t.Error(err)
		return
	}
	spew.Dump(b)
}
