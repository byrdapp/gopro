package thumbnail

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/jpeg"

	"github.com/davecgh/go-spew/spew"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/pkg/logger"
)

// TODO: format to jpeg: x-adobe-dmg && binary files
// https://github.com/h2non/bimg

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
func New(buf bytes.Buffer, filter ...Filter) (*Image, error) {
	img, err := decodeImg(buf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "decoding image from buffer")
	}

	cfg, ext, err := decodeImgCfg(buf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "decoding image config")
	}

	parseOpts := setDefaultParseOptions(filter...)
	return &Image{
		parseOptions: parseOpts,
		Extension:    ext,
		Info:         cfg,
		Image:        img,
		buf:          buf,
	}, nil
}

func byteReader(imageData []byte) *bytes.Reader {
	return bytes.NewReader(imageData)
}

func decodeImg(data []byte) (img image.Image, err error) {
	r := byteReader(data)
	img, err = imaging.Decode(r)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding raw image")
	}
	return img, err
}

func decodeImgCfg(data []byte) (cfg image.Config, ext string, err error) {
	r := byteReader(data)
	cfg, ext, err = image.DecodeConfig(r)
	if err != nil {
		return cfg, ext, errors.Wrap(err, "error decoding image config")
	}
	return cfg, ext, err
}

type parseOptions struct {
	width, height int
	filter        imaging.ResampleFilter
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

// If the format is anything else than JPEG, convert it...
func (img *Image) writeAsJPEG() (image.Image, error) {
	var err error
	ext, err := imaging.FormatFromExtension(img.Extension)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't format from extension: %s", img.Extension)
	}

	switch ext {
	case imaging.JPEG:
		log.Info("JPEG - no formatting needed")
		break
	case imaging.PNG:
		err = jpeg.Encode(&img.buf, img.Image, &jpeg.Options{}) // TODO: Quality
		break
	case imaging.GIF:
		err = jpeg.Encode(&img.buf, img.Image, &jpeg.Options{}) // TODO: Quality
		break
	default:
		err = errors.New("format is not supported yet")
		break
	}
	return img.Image, err
}

// Create thumbnail from imaging lib
func (img *Image) createThumbnail(opt parseOptions) *image.NRGBA {
	spew.Dump(opt)
	return imaging.Thumbnail(img.Image, opt.width, opt.height, opt.filter)
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

	parsedImg, err := img.writeAsJPEG()
	if err != nil {
		return nil, errors.Errorf("error writing file as jpeg: %s with ext: %s", err, img.Extension)
	}
	img.Image = parsedImg
	thumbnail := img.createThumbnail(img.parseOptions)

	buf := bytes.Buffer{}
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

// type ImageParser interface {
// 	Bytes() ([]byte, error)
// }

func (pImg *ParsedImage) Bytes() ([]byte, error) {
	err := jpeg.Encode(&pImg.buf, pImg.Thumbnail, nil)
	if err != nil {
		return nil, err
	}
	return pImg.buf.Bytes(), nil
}
