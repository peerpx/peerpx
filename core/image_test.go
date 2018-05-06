package core

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getImage() *Image {
	photoBytes, err := ioutil.ReadFile("../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	//  init mocked datastore
	DS = NewDatastoreMocked(photoBytes, nil)

	// get image from DS
	image, err := NewImageFromDataStore("fakeHAsh")
	if err != nil {
		panic(err)
	}
	return image
}

func TestNewImageFromDataStore(t *testing.T) {
	photoBytes, err := ioutil.ReadFile("../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	//  init mocked datastore
	DS = NewDatastoreMocked(photoBytes, nil)
	// get image from DS
	image, err := NewImageFromDataStore("fakeHAsh")
	if assert.NoError(t, err) {
		assert.Equal(t, 1000, image.Width())
		assert.Equal(t, 1270, image.Height())
	}
}

func TestImage_Resize(t *testing.T) {
	image := getImage()
	err := image.Resize(500, 200)
	if assert.NoError(t, err) {
		assert.Equal(t, 500, image.Width())
		assert.Equal(t, 200, image.Height())
	}
	// upscale (not allowed)
	err = image.Resize(600, 1000)
	assert.Equal(t, ErrImageUpscale, err)
}

func TestImage_ResizeToFit(t *testing.T) {
	image2 := getImage()
	err := image2.ResizeToFit(500, 500)
	if assert.NoError(t, err) {
		assert.Equal(t, 394, image2.Width())
		assert.Equal(t, 500, image2.Height())
	}
	// upscale (not allowed)
	err = image2.ResizeToFit(600, 1000)
	assert.Equal(t, ErrImageUpscale, err)
}
