package thumbnail

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

type Image struct {
	Extension    string
	Info         image.Config
	Image        image.Image
	buf          bytes.Buffer
	parseOptions parseOptions
}

/**
Constructor function to create new image processing. Filter is optional.
*/
func New(r io.Reader, filter ...Filter) (*Image, error) {
	img, err := decodeImg(r)
	if err != nil {
		return nil, err
	}
	cfg, ext, err := decodeImgCfg(r)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer

	parseOpts := setDefaultParseOptions(filter...)

	return &Image{
		parseOptions: parseOpts,
		Extension:    ext,
		Info:         cfg,
		Image:        img,
		buf:          buf,
	}, nil
}

type parseOptions struct {
	width, height int
	filter        imaging.ResampleFilter
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

// Sets default options and filter (Lanczos) if not specified otherwise
func setDefaultParseOptions(filter ...Filter) parseOptions {
	sampleFilter := Filter(imaging.Lanczos)

	if len(filter) > 0 {
		sampleFilter = filter[0]
	}
	return parseOptions{defaultWidth, defaultHeight, imaging.ResampleFilter(sampleFilter)}
}

func (img *Image) aboveThreshold(cfg image.Config) bool {
	return cfg.Width > widthResizeThreshold && cfg.Height > heightResizeThreshold
}

// Create thumbnail from imaging lib
func (img *Image) createThumbnail(opt parseOptions) *image.NRGBA {
	return imaging.Thumbnail(img.Image, opt.width, opt.height, opt.filter)
}

// If the format is anything else than JPEG, convert it...
func (img *Image) writeAsJPEG() (*Image, error) {
	var err error
	ext, err := imaging.FormatFromExtension(img.Extension)
	if err != nil {
		return nil, err
	}

	switch ext {
	case imaging.JPEG:
		break

	case imaging.PNG:
		err = jpeg.Encode(&img.buf, img.Image, &jpeg.Options{}) // TODO: Quality
		break

	default:
		err = errors.New("format is not supported yet")
		break
	}
	return img, err
}

/**
Resize image according to the default dimensions.
Default value for filter is imaging.Lanczos
*/
func (img *Image) Thumbnail() (*ParsedImage, error) {
	// is the img big enough to upload to byrd and therefore worth scaling?
	if !img.aboveThreshold(img.Info) {
		return nil, errors.New("image too small to upload to platform")
	}
	if img.Info.Width <= 0 || img.Info.Height <= 0 {
		return nil, errors.Errorf("dimensions cannot be negative or nil %v %v", img.Info.Width, img.Info.Height)
	}
	thumbnail := img.createThumbnail(img.parseOptions)

	var buf bytes.Buffer

	return &ParsedImage{
		Thumbnail: thumbnail,
		buf:       buf,
	}, nil
}

// ParsedImage contains the thumbnail properties
type ParsedImage struct {
	Thumbnail *image.NRGBA
	buf       bytes.Buffer
}

type ImageParser interface {
	Bytes() ([]byte, error)
}

func (pImg *ParsedImage) Bytes() ([]byte, error) {
	err := jpeg.Encode(&pImg.buf, pImg.Thumbnail, nil)
	if err != nil {
		return nil, err
	}
	return pImg.buf.Bytes(), nil
}
