package image

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getImage() []byte {
	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	return photoBytes
}

func TestImage(t *testing.T) {
	img, err := New(bytes.NewBuffer(getImage()))
	if assert.NoError(t, err) {
		assert.Equal(t, 1000, img.Width())
		assert.Equal(t, 1270, img.Height())

		// resize
		err = img.Resize(500, 200)
		if assert.NoError(t, err) {
			assert.Equal(t, 500, img.Width())
			assert.Equal(t, 200, img.Height())
		}
		// upscale (not allowed)
		err = img.Resize(600, 1000)
		assert.Equal(t, ErrUpscaleNotAllowed, err)

		// resize to fit
		err := img.ResizeToFit(250, 200)
		if assert.NoError(t, err) {
			assert.Equal(t, 250, img.Width())
			assert.Equal(t, 100, img.Height())
		}
		// upscale (not allowed)
		err = img.ResizeToFit(600, 1000)
		assert.Equal(t, ErrUpscaleNotAllowed, err)
	}
}
