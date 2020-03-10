package file

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/byrdapp/byrd-pro-api/pkg/conversion"
)

type File struct {
	file *os.File
}

func NewFileLtdRead(r io.Reader, limit int64) (*File, error) {
	// rd := io.LimitReader(r, limit)
	rd := io.LimitReader(r, 1000)
	b, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	return writeTmpFile(b)

}

// Read whole file at once
func NewFile(r io.Reader) (*File, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return writeTmpFile(b)
}

func NewEmptyFile() (*File, error) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix-*")
	if err != nil {
		return nil, err
	}

	return &File{file}, nil
}

// Read file buffered as scanner ! not tested !
func NewFileBuffer(r *bufio.Scanner) (*File, error) {
	var b []byte
	for r.Scan() {
		if err := r.Err(); err != nil {
			return nil, err
		}
	}
	return writeTmpFile(b)
}

func writeTmpFile(data []byte) (*File, error) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix-*")
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(file.Name(), data, 0777); err != nil {
		return nil, err
	}
	return &File{file}, nil
}

func (f *File) WriteFile(data []byte) (*File, error) {
	if err := ioutil.WriteFile(f.file.Name(), data, 0777); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *File) File() *os.File {
	return f.file
}

func (f *File) Close() error {
	return f.file.Close()
}

func (f *File) RemoveFile() error {
	return os.Remove(f.file.Name())
}

func (f *File) FileName() string {
	return f.file.Name()
}

func (f *File) FileStat() (os.FileInfo, error) {
	return f.file.Stat()
}

func (f *File) FileSize() (size float64, err error) {
	fInfo, err := f.file.Stat()
	if err != nil {
		return size, err
	}
	size = conversion.FileSizeBytesToFloat(int(fInfo.Size()))
	return size, err
}

func (f *File) EncodeExif(metaTag, value string) error {
	// Handle file types
	return nil
}

type Reader interface {
	Bytes() ([]byte, error)
	Read()
}

func (f *File) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, f.file)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
