package image

import (
	"io/ioutil"
	"testing"

	"github.com/peerpx/peerpx/services/datastore"
	"github.com/stretchr/testify/assert"
)

func getImage() *Image {
	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	//  init mocked datastore
	datastore.DS = datastore.NewMocked(photoBytes, nil)

	// get image from DS
	img, err := NewFromDataStore("fakeHAsh")
	if err != nil {
		panic(err)
	}
	return img
}

func TestNewFromDataStore(t *testing.T) {
	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	//  init mocked datastore
	datastore.DS = datastore.NewMocked(photoBytes, nil)
	// get image from DS
	img, err := NewFromDataStore("fakeHAsh")
	if assert.NoError(t, err) {
		assert.Equal(t, 1000, img.Width())
		assert.Equal(t, 1270, img.Height())
	}
}

func TestResize(t *testing.T) {
	img := getImage()
	err := img.Resize(500, 200)
	if assert.NoError(t, err) {
		assert.Equal(t, 500, img.Width())
		assert.Equal(t, 200, img.Height())
	}
	// upscale (not allowed)
	err = img.Resize(600, 1000)
	assert.Equal(t, ErrUpscale, err)
}

func TestResizeToFit(t *testing.T) {
	img := getImage()
	err := img.ResizeToFit(500, 500)
	if assert.NoError(t, err) {
		assert.Equal(t, 394, img.Width())
		assert.Equal(t, 500, img.Height())
	}
	// upscale (not allowed)
	err = img.ResizeToFit(600, 1000)
	assert.Equal(t, ErrUpscale, err)
}
