package controllers

import (
	"bytes"
	"fmt"
	"image"
	// jpeg
	_ "image/jpeg"
	// png
	_ "image/png"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"

	"image/jpeg"

	"github.com/disintegration/gift"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/toorop/peerpx/core"
	"github.com/toorop/peerpx/core/models"
)

// PhotoPostResponse is the response sent by PhotoPost ctrl
type PhotoPostResponse struct {
	Code    uint8
	Msg     string
	PhotoID string
}

// PhotoPost handle POST /api/v1.photo request
func PhotoPost(c echo.Context) error {
	// code:
	// 1 bad format (jpeg or png)
	response := PhotoPostResponse{}

	// get params
	// available
	// - name
	// - description TODO

	// get body -> photo
	photoBytes, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Infof("%v - controllers.PhotoPost - unable to read request body: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer c.Request().Body.Close()

	// check type
	mimeType := http.DetectContentType(photoBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		response.Code = 1
		response.Msg = fmt.Sprintf("only jpeg or png file are allowed, %s given", mimeType)
		return c.JSON(http.StatusBadRequest, response)
	}

	// resize && reencode
	reencodingNeeded := mimeType != "image/jpeg"
	pic, _, err := image.Decode(bytes.NewBuffer(photoBytes))
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to decode photo: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if pic.Bounds().Max.X > viper.GetInt("photo.maxWidth") || pic.Bounds().Max.Y > viper.GetInt("photo.maxHeight") {
		g := gift.New(
			gift.ResizeToFit(viper.GetInt("photo.maxWidth"), viper.GetInt("photo.maxHeight"), gift.LanczosResampling),
		)
		picResized := image.NewRGBA(g.Bounds(pic.Bounds()))
		g.Draw(picResized, pic)
		pic = picResized
		reencodingNeeded = true
	}
	// re-encoding
	if reencodingNeeded {
		buf := bytes.NewBuffer([]byte{})
		options := jpeg.Options{Quality: 100}
		if err = jpeg.Encode(buf, pic, &options); err != nil {
			log.Errorf("%v - controllers.PhotoPost - unable to reencode photo: %v", c.RealIP(), err)
			return c.NoContent(http.StatusInternalServerError)
		}
		photoBytes, err = ioutil.ReadAll(buf)
		if err != nil {
			log.Errorf("%v - controllers.PhotoPost - unable to read buffer when reencoding photo: %v", c.RealIP(), err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	photo := models.Photo{}

	// get hash
	photo.Hash, err = core.GetHash(photoBytes)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to get photo hash: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// get final size
	pic, _, err = image.Decode(bytes.NewBuffer(photoBytes))
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to decode photo to get final size: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	photo.Width = uint32(pic.Bounds().Max.X)
	photo.Height = uint32(pic.Bounds().Max.Y)

	// URL
	if viper.GetBool("http.tlsEnabled") {
		photo.URL = fmt.Sprintf("https://%s/api/v1/photo/%s/raw", viper.GetString("hostname"), photo.Hash)
	} else {
		photo.URL = fmt.Sprintf("http://%s/api/v1/photo/%s/raw", viper.GetString("hostname"), photo.Hash)
	}

	// save in datastore
	if err = core.DS.Put(photo.Hash, photoBytes); err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to store photo in datastore: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// create entry in DB
	if err = photo.Create(); err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to photo.Create: %v", c.RealIP(), err)
		// remove photo from datastore
		if err = core.DS.Delete(photo.Hash); err != nil {
			log.Errorf("%v - controllers.PhotoPost - unable to remove photo %s datastore: %v", c.RealIP(), photo.Hash, err)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	// return response
	response.PhotoID = photo.Hash
	return c.JSON(http.StatusCreated, response)

}

// PhotoSearch API ctrl for searching photos
func PhotoSearch(c echo.Context) error {
	log.Print("toto")
	log.Error("toto")
	return c.String(http.StatusOK, "TODO")
}