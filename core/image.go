package core

import (
	"bytes"
	"errors"
	imageStd "image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"

	"github.com/disintegration/gift"
)

// Image is used for image manipulation
type Image struct {
	image  imageStd.Image
	format string
}

var (
	ErrImageUpscale = errors.New("upscaling is not allowed")
)

// NewImageFromDataStore return an instantiated Image fetched from datastore
func NewImageFromDataStore(hash string) (image *Image, err error) {
	var imageBytes []byte
	image = new(Image)
	imageBytes, err = DS.Get(hash)
	if err != nil {
		return
	}
	image.image, image.format, err = imageStd.Decode(bytes.NewBuffer(imageBytes))
	return
}

// NewImageFromBytes returns image from bytes slice
func NewImageFromBytes(b []byte) (image *Image, err error) {
	image = new(Image)
	image.image, image.format, err = imageStd.Decode(bytes.NewBuffer(b))
	return
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
		return ErrImageUpscale
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
		return ErrImageUpscale
	}
	g := gift.New(
		gift.ResizeToFit(width, height, gift.LanczosResampling),
	)
	resized := imageStd.NewRGBA(g.Bounds(i.image.Bounds()))
	g.Draw(resized, i.image)
	i.image = resized
	return nil
}
