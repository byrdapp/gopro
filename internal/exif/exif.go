package exif

// Output represents the final decoded EXIF data from an image
type Output struct {
	Date            int64   `json:"date,omitempty"`
	Lat             float64 `json:"lat,omitempty"`
	Lng             float64 `json:"lng,omitempty"`
	Copyright       string  `json:"copyright,omitempty"`
	Model           string  `json:"model,omitempty"`
	PixelXDimension int     `json:"pixelXDimension,omitempty"`
	PixelYDimension int     `json:"pixelYDimension,omitempty"`
	MediaSize       float64 `json:"mediaSize,omitempty"`
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

// type ExifReader interface {
// 	VideoFile() (*Output, error)
// 	ReadImage(r io.Reader) (*Output, error)
// 	imageReader(r io.Reader) (*Output, error)
// }
