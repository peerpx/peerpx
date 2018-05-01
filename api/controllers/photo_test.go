package controllers

import (
	"io/ioutil"
	"testing"

	"bytes"
	"net/http/httptest"

	"net/http"

	"encoding/json"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestPhotoPost(t *testing.T) {
	// init viper
	viper.Set("photo.maxWidth", 100)
	viper.Set("photo.maxHeight", 100)

	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/robin.jpg")
	if err != nil {
		panic(err)
	}
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", bytes.NewBuffer(photoBytes))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, PhotoPost(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		assert.NoError(t, err)
		var response PhotoPostResponse
		err = json.Unmarshal(resp, &response)
		assert.NoError(t, err)
		assert.Equal(t, "2DJLYuo9ky9CfThuGK2DU82dvENtJr8BzX7kmGkoad4J", response.PhotoID)
	}

}

func TestPhotoPostNotAPhoto(t *testing.T) {
	photoBytes, err := ioutil.ReadFile("../../etc/samples/photos/not-a-photo.jpg")
	if err != nil {
		panic(err)
	}
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", bytes.NewBuffer(photoBytes))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, PhotoPost(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		assert.NoError(t, err)
		var response PhotoPostResponse
		err = json.Unmarshal(resp, &response)
		assert.NoError(t, err)
		assert.Equal(t, uint8(1), response.Code)
	}

}
