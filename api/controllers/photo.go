package controllers

import (
	// jpeg
	_ "image/jpeg"
	// png
	_ "image/png"
	"net/http"

	"github.com/labstack/echo"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/core"
	"github.com/peerpx/peerpx/core/models"
	log "github.com/sirupsen/logrus"
	//"encoding/json"
	//"fmt"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"strconv"

	"github.com/spf13/viper"
)

// PhotoPostResponse is the response sent by PhotoPost ctrl
// TODO exif
type PhotoPostResponse struct {
	Code       uint8        `json:"code"`
	Msg        string       `json:"msg"`
	PhotoProps models.Photo `json:"photoProps"`
}

// PhotoPost handle POST /api/v1.photo request
func PhotoPost(c echo.Context) error {
	// code:
	// 1 bad photo format (jpeg or png)
	// 2 bad data (not valid photo struct/object)
	// 3 bad file
	response := PhotoPostResponse{}

	// init photo
	photo := models.Photo{}

	form, err := c.MultipartForm()
	if err != nil {
		panic(err)
	}

	// get photo props
	data := form.Value["data"]

	if len(data) != 1 {
		log.Infof("%v - controllers.PhotoPost - bad request len(data) = %d, 1 expected", c.RealIP(), len(data))
		response.Code = 2
		response.Msg = fmt.Sprintf("bad data lenght")
		return c.JSON(http.StatusBadRequest, response)
	}

	// TODO verifier les props requises

	if err := json.Unmarshal([]byte(data[0]), &photo); err != nil {
		log.Infof("ERR %v", err)
		response.Code = 2
		response.Msg = fmt.Sprintf("bad data: '%s' given", data)
		return c.JSON(http.StatusBadRequest, response)
	}

	// get photo file
	photoFile := form.File["file"]
	if len(photoFile) != 1 {
		log.Infof("%v - controllers.PhotoPost - bad request len(photoFile) = %d, 1 expected", c.RealIP(), len(photoFile))
		response.Code = 3
		response.Msg = fmt.Sprintf("bad photoFile lenght")
		return c.JSON(http.StatusBadRequest, response)
	}

	// get body -> photo
	fd, err := photoFile[0].Open()
	if err != nil {
		log.Infof("%v - controllers.PhotoPost - unable to read photoFile: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer fd.Close()

	photoBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Infof("%v - controllers.PhotoPost - unable to ioutil.ReadAll(fd): %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// check type
	mimeType := http.DetectContentType(photoBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		log.Infof("%v - controllers.PhotoPost - bad file type: %s", c.RealIP(), mimeType)
		response.Code = 1
		response.Msg = fmt.Sprintf("only jpeg or png file are allowed, %s given", mimeType)
		return c.JSON(http.StatusBadRequest, response)
	}

	// resize && re-encode
	image, err := core.NewImageFromBytes(photoBytes)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to core.NewImageFromBytes(): %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if image.Width() > viper.GetInt("photo.maxWidth") || image.Height() > viper.GetInt("photo.maxHeight") {
		err = image.ResizeToFit(viper.GetInt("photo.maxWidth"), viper.GetInt("photo.maxHeight"))
		if err != nil && err != core.ErrImageUpscale {
			log.Errorf("%v - controllers.PhotoPost - unable to image.ResizeToFit(): %v", c.RealIP(), err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	//
	photoBytes, err = image.JPEG(100)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to image.JPEG(): %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// get hash
	photo.Hash, err = core.GetHash(photoBytes)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to get photo hash: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	//  size
	photo.Width = uint32(image.Width())
	photo.Height = uint32(image.Height())

	// URL
	// TODO a supprimer ?
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
	response.PhotoProps = photo
	return c.JSON(http.StatusCreated, response)
}

// PhotoGetProperties returns PhotoProperties
func PhotoGetProperties(c echo.Context) error {
	// get ID -> hash
	hash := c.Param("id")
	// get photo
	photo, err := models.PhotoGetByHash(hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("%v - controllers.PhotoGetProperties - unable to models.PhotoGetByHash(%s): %v", c.RealIP(), hash, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, photo)
}

// PhotoGet return a photo
func PhotoGet(c echo.Context) error {
	// get hash & size
	hash := c.Param("id")
	size := c.Param("size")
	// osef de size for now
	_ = size

	// get photo from data store
	photoBytes, err := core.DS.Get(hash)
	if err != nil {
		if err == core.ErrNotFoundInDatastore {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("%v - controllers.PhotoGet - unable to get %s from datastore: %v", c.RealIP(), hash, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Blob(http.StatusOK, "image/jpeg ", photoBytes)
}

// PhotoPutResponse
type PhotoPutResponse struct {
	Code  uint8         `json:"code"`
	Photo *models.Photo `json:"photo"`
}

// PhotoPut alter photo properties
func PhotoPut(c echo.Context) error {

	// read body
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		log.Errorf("%v - controllers.PhotoPut - read request body failed: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// -> photoNew
	var photoNew models.Photo
	if err := json.Unmarshal(bodyBytes, &photoNew); err != nil {
		log.Errorf("%v - controllers.PhotoPut - unmarshall photoNew failed : %v", c.RealIP(), err)
		return c.NoContent(http.StatusBadRequest)
	}

	// init response
	response := PhotoPutResponse{}

	// validate
	if status := photoNew.Validate(); status != 0 {
		response.Code = status
		return c.JSON(http.StatusBadRequest, response)

	}

	// get photo props ->  photoOri
	photoOri, err := models.PhotoGetByHash(photoNew.Hash)
	switch err {
	case gorm.ErrRecordNotFound:
		return c.NoContent(http.StatusNotFound)
	case nil:
	default:
		log.Errorf("%v - controllers.PhotoPut - models.PhotoGetByHash(%s) failed : %v", c.RealIP(), photoNew.Hash, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// PhotoNew -> PhotoOri (update)
	photoOri.Name = photoNew.Name
	photoOri.Description = photoNew.Description
	photoOri.Camera = photoNew.Camera
	photoOri.Lens = photoNew.Lens
	photoOri.FocalLength = photoNew.FocalLength
	photoOri.Iso = photoNew.Iso
	photoOri.ShutterSpeed = photoNew.ShutterSpeed
	photoOri.Aperture = photoNew.Aperture
	// TODO Category     Category
	photoOri.Location = photoNew.Location
	photoOri.Privacy = photoNew.Privacy
	photoOri.Latitude = photoNew.Latitude
	photoOri.Longitude = photoNew.Longitude
	photoOri.TakenAt = photoNew.TakenAt
	photoOri.Nsfw = photoNew.Nsfw
	// TODO LicenceType  Licence

	// photo.Update -> DB
	if err = photoOri.Update(); err != nil {
		log.Errorf("%v - controllers.PhotoPut - photoOri.Update() failed : %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// return photo
	response.Photo = &photoOri
	return c.JSON(http.StatusOK, response)
}

// PhotoDel delete a photo
func PhotoDel(c echo.Context) error {
	// get hash
	hash := c.Param("id")
	if err := models.PhotoDeleteByHash(hash); err != nil {
		log.Errorf("%v - controllers.PhotoGet - unable to delete photo %s: %v", c.RealIP(), hash, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// PhotoResize returns resized photo
func PhotoResize(c echo.Context) error {
	var width, height int
	var err error
	// hauteur ou largeur
	widthStr := c.Param("width")
	heightStr := c.Param("height")

	if widthStr == "" {
		width = 0
	} else {
		width, err = strconv.Atoi(widthStr)
		if err != nil {
			log.Errorf("%v - controllers.PhotoResize - unable to strconv.Atoi(%s): %v", c.RealIP(), widthStr, err)
			return c.NoContent(http.StatusBadRequest)
		}
	}

	if heightStr == "" {
		height = 0
	} else {
		height, err = strconv.Atoi(heightStr)
		if err != nil {
			log.Errorf("%v - controllers.PhotoResize - unable to strconv.Atoi(%s): %v", c.RealIP(), heightStr, err)
			return c.NoContent(http.StatusBadRequest)
		}
	}
	if height == 0 && width == 0 {
		log.Errorf("%v - controllers.PhotoResize - height == width == 0", c.RealIP())
		return c.NoContent(http.StatusBadRequest)
	}

	image, err := core.NewImageFromDataStore(c.Param("id"))
	if err != nil {
		log.Errorf("%v - controllers.PhotoResize - unable to core.NewImageFromDataStore(%s): %v", c.RealIP(), c.Param("id"), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = image.Resize(width, height); err != nil {
		log.Errorf("%v - controllers.PhotoResize - unable to image.ResizeToFit(%d, %d): %v", c.RealIP(), width, height, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	b, err := image.JPEG(100)
	if err != nil {
		log.Errorf("%v - controllers.PhotoResize - unable to image.JPEG(): %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Blob(http.StatusOK, "image/jpeg", b)
}

// PhotoSearchResponse response structure for PhotoSearch
type PhotoSearchResponse struct {
	Total  int
	Limit  int
	Offset int
	Data   []models.Photo
}

// PhotoSearch return an array of photos regarding the optionnals search params (TMP)
func PhotoSearch(c echo.Context) error {
	//TODO: take account of optional params
	photos, err := models.PhotoList()
	if err != nil {
		log.Errorf("%v - controllers.PhotoSearch - unable to list photos: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	response := PhotoSearchResponse{
		Total:  len(photos),
		Limit:  0,
		Offset: 0,
		Data:   photos,
	}
	return c.JSON(http.StatusOK, response)
}
