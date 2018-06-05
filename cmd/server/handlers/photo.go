package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"strconv"

	"strings"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/entities/photo"
	"github.com/peerpx/peerpx/pkg/hasher"
	"github.com/peerpx/peerpx/pkg/image"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/datastore"
	log "github.com/sirupsen/logrus"
)

// PhotoPostResponse is the response sent by PhotoPost ctrl
// TODO exif
type PhotoPostResponse struct {
	Code       uint8       `json:"code"`
	Msg        string      `json:"msg"`
	PhotoProps photo.Photo `json:"photoProps"`
}

// PhotoPost handle POST /api/v1.photo request
func PhotoPost(c echo.Context) error {
	// code:
	// 1 bad photo format (jpeg or png)
	// 2 bad data (not valid photo struct/object)
	// 3 bad file
	// 4 duplicate
	// 404 not found
	// 500 internal server error
	response := PhotoPostResponse{}

	// init photo
	phot := photo.Photo{}

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

	if err := json.Unmarshal([]byte(data[0]), &phot); err != nil {
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
		response.Code = 500
		response.Msg = "unable to read form file"
		return c.JSON(http.StatusInternalServerError, response)
	}
	defer fd.Close()

	photoBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Infof("%v - controllers.PhotoPost - unable to ioutil.ReadAll(fd): %v", c.RealIP(), err)
		response.Code = 500
		response.Msg = "unable to read form file #2"
		return c.JSON(http.StatusInternalServerError, response)
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
	img, err := image.NewFromBytes(photoBytes)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to core.NewImageFromBytes(): %v", c.RealIP(), err)
		response.Code = 500
		response.Msg = "unable to create new image"
		return c.JSON(http.StatusInternalServerError, response)
	}
	if img.Width() > config.GetIntDefault("photo.maxWidth", 2000) || img.Height() > config.GetIntDefault("photo.maxHeight", 2000) {
		err = img.ResizeToFit(config.GetIntDefault("photo.maxWidth", 2000), config.GetIntDefault("photo.maxHeight", 2000))
		if err != nil && err != image.ErrUpscaleNotAllowed {
			log.Errorf("%v - controllers.PhotoPost - unable to img.ResizeToFit(): %v", c.RealIP(), err)
			response.Code = 500
			response.Msg = "unable to resize image"
			return c.JSON(http.StatusInternalServerError, response)
		}
	}

	//
	photoBytes, err = img.JPEG(100)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to img.JPEG(): %v", c.RealIP(), err)
		response.Code = 500
		response.Msg = "unable to convert image to JPEG"
		return c.JSON(http.StatusInternalServerError, response)
	}

	// get hash
	phot.Hash, err = hasher.GetHash(photoBytes)
	if err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to get photo hash: %v", c.RealIP(), err)
		response.Code = 500
		response.Msg = "unable to get hash from image"
		return c.JSON(http.StatusInternalServerError, response)
	}

	//  size
	phot.Width = uint32(img.Width())
	phot.Height = uint32(img.Height())

	// URL
	// TODO a supprimer ?
	if config.GetBool("http.tlsEnabled") {
		phot.URL = fmt.Sprintf("https://%s/api/v1/photo/%s/raw", config.GetStringP("hostname"), phot.Hash)
	} else {
		phot.URL = fmt.Sprintf("http://%s/api/v1/photo/%s/raw", config.GetStringP("hostname"), phot.Hash)
	}

	// save in datastore
	if err = datastore.Put(phot.Hash, photoBytes); err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to store photo in datastore: %v", c.RealIP(), err)
		response.Code = 500
		response.Msg = "unable to save image to datastore"
		return c.JSON(http.StatusInternalServerError, response)
	}

	// create entry in DB
	if err = phot.Create(); err != nil {
		log.Errorf("%v - controllers.PhotoPost - unable to photo.Create: %v", c.RealIP(), err)
		// remove photo from datastore
		if !strings.HasPrefix(err.Error(), "UNIQUE") {
			if err = datastore.Delete(phot.Hash); err != nil {
				log.Errorf("%v - controllers.PhotoPost - unable to remove photo %s datastore: %v", c.RealIP(), phot.Hash, err)
			}
			response.Code = 4
			response.Msg = "duplicate"
			return c.JSON(http.StatusBadRequest, response)
		}
		response.Code = 500
		response.Msg = "unable to save image to DB"
		return c.JSON(http.StatusInternalServerError, response)

	}

	// return response
	response.PhotoProps = phot
	return c.JSON(http.StatusCreated, response)
}

// PhotoGetProperties returns PhotoProperties
func PhotoGetProperties(c echo.Context) error {
	// get ID -> hash
	hash := c.Param("id")
	// get photo
	phot, err := photo.GetByHash(hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("%v - controllers.PhotoGetProperties - unable to models.PhotoGetByHash(%s): %v", c.RealIP(), hash, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, phot)
}

// PhotoGet return a photo
func PhotoGet(c echo.Context) error {
	// get hash & size
	hash := c.Param("id")
	size := c.Param("size")
	// osef de size for now
	_ = size

	// get photo from data store
	photoBytes, err := datastore.Get(hash)
	if err != nil {
		if err == datastore.ErrNotFound {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("%v - controllers.PhotoGet - unable to get %s from datastore: %v", c.RealIP(), hash, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Blob(http.StatusOK, "image/jpeg ", photoBytes)
}

// PhotoPutResponse
type PhotoPutResponse struct {
	Code  uint8        `json:"code"`
	Photo *photo.Photo `json:"photo"`
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
	var photoNew photo.Photo
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
	photoOri, err := photo.GetByHash(photoNew.Hash)
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
	if err := photo.DeleteByHash(hash); err != nil {
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

	imgBytes, err := datastore.Get(c.Param("id"))
	if err != nil {
		log.Errorf("%v - controllers.PhotoResize - datastore.get(%s) failed: %v", c.RealIP(), c.Param("id"), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	img, err := image.New(bytes.NewBuffer(imgBytes))
	if err != nil {
		log.Errorf("%v - controllers.PhotoResize - unable to core.NewImageFromDataStore(%s): %v", c.RealIP(), c.Param("id"), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err = img.Resize(width, height); err != nil {
		log.Errorf("%v - controllers.PhotoResize - unable to img.ResizeToFit(%d, %d): %v", c.RealIP(), width, height, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	b, err := img.JPEG(100)
	if err != nil {
		log.Errorf("%v - controllers.PhotoResize - unable to img.JPEG(): %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.Blob(http.StatusOK, "image/jpeg", b)
}

// PhotoSearchResponse response structure for PhotoSearch
type PhotoSearchResponse struct {
	Total  int
	Limit  int
	Offset int
	Data   []photo.Photo
}

// PhotoSearch return an array of photos regarding the optionnals search params (TMP)
func PhotoSearch(c echo.Context) error {
	//TODO: take account of optional params
	photos, err := photo.List()
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
