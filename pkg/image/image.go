package image

import (
	"bytes"
	"errors"
	imageStd "image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"

	"io"

	"github.com/disintegration/gift"
)

// Image is used for image manipulation
type Image struct {
	image  imageStd.Image
	format string
}

var (
	ErrUpscaleNotAllowed = errors.New("upscaling is not allowed")
)

// New returns image from io.Reader
func New(r io.Reader) (image *Image, err error) {
	image = new(Image)
	image.image, image.format, err = imageStd.Decode(r)
	return
}

// NewFromBytes returns image from bytes slice
func NewFromBytes(b []byte) (image *Image, err error) {
	return New(bytes.NewBuffer(b))
}

// Width returns image width
func (i *Image) Width() int {
	return i.image.Bounds().Max.X
}

// Height return image height
func (i *Image) Height() int {
	return i.image.Bounds().Max.Y
}

// JPEG return image as jpeg
func (i *Image) JPEG(quality int) ([]byte, error) {
	var err error
	buf := bytes.NewBuffer([]byte{})
	options := jpeg.Options{Quality: quality}
	if err = jpeg.Encode(buf, i.image, &options); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(buf)
}

// Resize resize image
// Warning: upscaling not allowed
func (i *Image) Resize(width, height int) error {
	if width > i.Width() || height > i.Height() {
		return ErrUpscaleNotAllowed
	}
	g := gift.New(
		gift.Resize(width, height, gift.LanczosResampling),
	)
	resized := imageStd.NewRGBA(g.Bounds(i.image.Bounds()))
	g.Draw(resized, i.image)
	i.image = resized
	return nil
}

// ResizeToFit resize image to fit width,height
// Warning: upscaling not allowed
func (i *Image) ResizeToFit(width, height int) error {
	if width > i.Width() || height > i.Height() {
		return ErrUpscaleNotAllowed
	}
	g := gift.New(
		gift.ResizeToFit(width, height, gift.LanczosResampling),
	)
	resized := imageStd.NewRGBA(g.Bounds(i.image.Bounds()))
	g.Draw(resized, i.image)
	i.image = resized
	return nil
}
