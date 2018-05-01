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

	//  get hash
	response.PhotoID, err = core.GetHash(photoBytes)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to get photo hash: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
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
	// reencoding
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

	// save

	// return
	return c.JSON(http.StatusCreated, response)

}

// PhotoSearch API ctrl for searching photos
func PhotoSearch(c echo.Context) error {
	log.Print("toto")
	log.Error("toto")
	return c.String(http.StatusOK, "TODO")
}
