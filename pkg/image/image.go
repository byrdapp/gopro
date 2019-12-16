package image

import (
	"image"
	"io"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/pkg/logger"
)

type ParsedImage struct {
	Extension string       `json:"extension"`
	Config    image.Config `json:"config"`
	Img       image.Image  `json:"img"`
}

var log = logger.NewLogger()

const (
	defaultWidth, defaultHeight                 = 640, 640
	widthResizeThreshold, heightResizeThreshold = 720, 720
)

func NewPreviewImage(r io.Reader) (*ParsedImage, error) {
	img, err := decodeImg(r)
	if err != nil {
		return nil, err
	}
	cfg, ext, err := decodeImgCfg(r)
	if err != nil {
		return nil, err
	}
	return &ParsedImage{
		Extension: ext,
		Config:    cfg,
		Img:       img,
	}, nil
}

func decodeImg(r io.Reader, opts ...imaging.DecodeOption) (img image.Image, err error) {
	img, err = imaging.Decode(r, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding image")
	}

	return img, err
}

func decodeImgCfg(r io.Reader) (cfg image.Config, ext string, err error) {
	cfg, ext, err = image.DecodeConfig(r)
	if err != nil {
		return cfg, "", errors.Wrap(err, "error decoding image")
	}
	return cfg, ext, err
}

func (img *ParsedImage) aboveThreshold(cfg image.Config) bool {
	return cfg.Width > widthResizeThreshold && cfg.Height > heightResizeThreshold
}

/**
Resize image according to the dimensions.
Default value for filter is imaging.Lanczos
*/
func (img *ParsedImage) Resize(dimX, dimY int, filter ...imaging.ResampleFilter) (*image.NRGBA, error) {
	// is the img big enough to upload to byrd and therefore worth scaling?
	if !img.aboveThreshold(img.Config) {
		return nil, errors.New("image too small to upload to platform")
	}
	if dimX <= 0 || dimY <= 0 {
		return nil, errors.Errorf("dimensions cannot be negative or nil %v %v", dimX, dimY)
	}
	opt := imaging.Lanczos
	if len(filter) > 0 {
		opt = filter[0]
	}
	return imaging.Resize(img.Img, dimX, dimY, opt), nil
}
