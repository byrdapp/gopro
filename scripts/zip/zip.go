package zip

import (
	"archive/zip"
	"compress/flate"
	"io"
	"os"
	"path/filepath"

	"github.com/google/martian/log"
)

func main() {
	p, err := filepath.Abs("internal/exif/video/video/in.mp4")
	if err != nil {
		log.Errorf("%s", err)
	}
	err = writeZip("test.zip", p)
	if err != nil {
		log.Errorf("%s", err)
	}
}

func writeZip(zipName string, file string) error {
	newZipFile, err := os.Create("test.zip")
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestSpeed)
	})

	// Add files to zip
	if err = AddFileToZip(zipWriter, file); err != nil {
		return err
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filename
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
