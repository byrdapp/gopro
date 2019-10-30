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
	ExifError       *ExifError
	// MediaFormat     string  `json:"mediaFormat,omitempty"`
}

type ExifError struct {
	Error map[string]interface{} `json:"error"`
}

// type ExifReader interface {
// 	VideoFile() (*Output, error)
// 	ReadImage(r io.Reader) (*Output, error)
// 	imageReader(r io.Reader) (*Output, error)
// }
