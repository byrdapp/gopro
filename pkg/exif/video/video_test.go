package video

import (
	"bufio"
	"io"
	"os"
	"testing"
)

func TestVideoExif(t *testing.T) {
	t.Run("read video metadata", func(t *testing.T) {
		// pr, pw := io.Pipe()
		// if err := aws.ParseCredentials(); err != nil {
		// 	t.Error(err)
		// 	return
		// }
		// mat, err := aws.GetTestMaterial("videos", "in.mp4")
		// if err != nil {
		// 	t.Error(err)
		// 	return
		// }

		// video, err := video.ReadVideo(r)
		// if err != nil {
		// 	t.Log(err)
		// }
		//
		// defer func() {
		// 	video.File.RemoveFile()
		// 	video.File.Close()
		// }()
		//
		// out := video.CreateVideoExifOutput()
		// if err != nil {
		// 	t.Log(err)
		// }
		//
		// spew.Dump(out)

	})
}

func readAsIoReader(fileName string) io.Reader {
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	return bufio.NewReader(f)
}
