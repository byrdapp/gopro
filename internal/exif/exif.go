package exif

// type MediaDimension int

// Output represents the final decoded EXIF data from an image
type Output struct {
	Date            int64             `json:"date,omitempty"`
	Lat             float64           `json:"lat,omitempty"`
	Lng             float64           `json:"lng,omitempty"`
	Copyright       string            `json:"copyright,omitempty"`
	Model           string            `json:"model,omitempty"`
	PixelXDimension int               `json:"pixelXDimension,omitempty"`
	PixelYDimension int               `json:"pixelYDimension,omitempty"`
	MediaSize       float64           `json:"mediaSize,omitempty"`
	ExifErrors      map[string]string `json:"errors,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

// adds an object to the output JSON that displays missing exif data
func (o *Output) MissingExif(errType string, err error) {
	o.ExifErrors[errType] = err.Error()
}

// type ExifReader interface {
// 	VideoFile() (*Output, error)
// 	ReadImage(r io.Reader) (*Output, error)
// 	imageReader(r io.Reader) (*Output, error)
// }
