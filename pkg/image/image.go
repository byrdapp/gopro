package image

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/pkg/logger"
)

var log = logger.NewLogger()

type Filter imaging.ResampleFilter

const (
	defaultWidth, defaultHeight                 = 640, 640
	widthResizeThreshold, heightResizeThreshold = 720, 720
)

type options struct {
	width, height int
	filter        imaging.ResampleFilter
}

type ParsedImage struct {
	options   options
	buf       bytes.Buffer
	Extension string
	Config    image.Config
	Img       image.Image
}

/**
Constructor function to create new image processing
*/
func NewPreviewProcessing(r io.Reader, filter ...Filter) (*ParsedImage, error) {
	img, err := decodeImg(r)
	if err != nil {
		return nil, err
	}
	cfg, ext, err := decodeImgCfg(r)
	if err != nil {
		return nil, err
	}
	sampleFilter := Filter(imaging.Lanczos)

	if len(filter) > 0 {
		sampleFilter = filter[0]
	}

	opt := options{cfg.Height, cfg.Width, imaging.ResampleFilter(sampleFilter)}

	return &ParsedImage{
		Extension: ext,
		Config:    cfg,
		Img:       img,
		options:   opt,
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

func (img *ParsedImage) writeAsJPEG() (*ParsedImage, error) {
	var err error
	ext, err := imaging.FormatFromExtension(img.Extension)
	if err != nil {
		return nil, err
	}

	switch ext {
	case imaging.JPEG:
		break

	case imaging.PNG:
		err = jpeg.Encode(&img.buf, img.Img, &jpeg.Options{}) // TODO: Quality
		break

	default:
		err = errors.New("format is not supported yet")
		break
	}
	return img, err
}

type Preview struct {
	Thumbnail []byte `json:"image"`
	Error     error  `json:"error"`
}

/**
Resize image according to the dimensions.
Default value for filter is imaging.Lanczos
*/
func (img *ParsedImage) Thumbnail() (*Preview, error) {
	// is the img big enough to upload to byrd and therefore worth scaling?
	if !img.aboveThreshold(img.Config) {
		return nil, errors.New("image too small to upload to platform")
	}
	if img.options.width <= 0 || img.options.height <= 0 {
		return nil, errors.Errorf("dimensions cannot be negative or nil %v %v", img.options.width, img.options.height)
	}
	thumb, err := img.createThumbnail(img.options)

	return &Preview{
		Thumbnail: nil,
		Error:     err,
	}, nil
}

func (img *ParsedImage) createThumbnail(opt options) (*image.NRGBA, error) {
	return imaging.Thumbnail(img.Img, opt.width, opt.height, opt.filter), nil
}

func (img *ParsedImage) parseImage() ([]byte, error) {
	err := jpeg.Encode(&img.buf, img.Img, nil)
	if err != nil {
		return nil, err
	}
	return img.buf.Bytes(), nil
}
