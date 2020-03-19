package compression

import (
	"archive/zip"
	"compress/flate"
	"io"
)

type ZipFile struct {
	// file   *os.File
	fileName string
	writer   *zip.Writer
}

type WriteCloser interface {
	Close() error
	Write() (int64, error)
}

func NewZip(w io.Writer, fileName string) (*ZipFile, error) {
	writer := zip.NewWriter(w)
	registerWriter(writer)
	return &ZipFile{
		fileName,
		writer,
	}, nil
}

func registerWriter(zipWriter *zip.Writer) {
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestSpeed)
	})
}

func (zf *ZipFile) Write(fileName string) (io.Writer, error) {
	// Get the file information
	// header.Name = info.Name()
	// header.Method = zip.Deflate
	return zf.writer.Create(fileName)
	// return zf.writer.CreateHeader(header)
}

func (zf *ZipFile) Close() error {
	return zf.writer.Close()
}
