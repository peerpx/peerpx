package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/entities/photo"
	"github.com/peerpx/peerpx/pkg/hasher"
	"github.com/peerpx/peerpx/pkg/image"
	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/log"
)

// PhotoCreate handle POST /api/v1.photo request
// response.Code:
// unsupportedPhotoFormat: bad photo format (jpeg or png)
// badData: bad data (not valid photo struct/object)
// badFile: bad file
// duplicate: duplicate
func PhotoCreate(ac echo.Context) error {
	c := ac.(*middlewares.AppContext)
	response := NewApiResponse(c.UUID)

	// get multipart
	form, err := c.MultipartForm()
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - c.MultipartForm() failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "reqNotMultipart"
		return c.JSON(response.HttpStatus, response)
	}

	// get photo properties
	photoProperties := form.Value["properties"]
	if len(photoProperties) != 1 {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - len(form.Value[properties]) != 1", c.RealIP(), response.UUID)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "reqBadMultipartProperties"
		return c.JSON(response.HttpStatus, response)
	}

	// TODO verifier les props

	// unmarshall photo
	p := new(photo.Photo)
	if err := json.Unmarshal([]byte(photoProperties[0]), p); err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - umarshall form.Value[properties] failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "reqBadPhotoProperties"
		return c.JSON(response.HttpStatus, response)
	}

	// get raw photo ("file")
	photoFile := form.File["file"]
	if len(photoFile) != 1 {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - len(form.File[file]) != 1", c.RealIP(), response.UUID)
		log.Error(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "reqBadMultipartFile"
		return c.JSON(response.HttpStatus, response)
	}

	// open photo file
	fd, err := photoFile[0].Open()
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - photoFile[0].Open() failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "openFormFileFailed"
		return c.JSON(response.HttpStatus, response)
	}
	defer fd.Close()

	// read photo file
	photoBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - ioutil.ReadAll(photoFile) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "readFormFileFailed"
		return c.JSON(response.HttpStatus, response)
	}

	// check mime type
	mimeType := http.DetectContentType(photoBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - %s is not supported for photo", c.RealIP(), response.UUID, mimeType)
		log.Info(response.Message)
		response.HttpStatus = http.StatusBadRequest
		response.Code = "unsupportedPhotoFormat"
		return c.JSON(response.HttpStatus, response)
	}

	// resize && re-encode
	img, err := image.NewFromBytes(photoBytes)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - image.NewFromBytes(photoBytes) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "imageNewFailed"
		return c.JSON(response.HttpStatus, response)
	}

	if img.Width() > config.GetIntDefault("photo.maxWidth", 2000) || img.Height() > config.GetIntDefault("photo.maxHeight", 2000) {
		err = img.ResizeToFit(config.GetIntDefault("photo.maxWidth", 2000), config.GetIntDefault("photo.maxHeight", 2000))
		if err != nil && err != image.ErrUpscaleNotAllowed {
			response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - img.ResizeToFit failed: %v", c.RealIP(), response.UUID, err)
			log.Error(response.Message)
			response.HttpStatus = http.StatusInternalServerError
			response.Code = "resizeFailed"
			return c.JSON(response.HttpStatus, response)
		}
	}

	// To JPEG
	photoBytes, err = img.JPEG(100)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - img.JPEG failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "conversionJpegFailed"
		return c.JSON(response.HttpStatus, response)
	}

	// get hash
	p.Hash, err = hasher.GetHash(photoBytes)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - hasher.GetHash(photoBytes) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "conversionJpegFailed"
		return c.JSON(response.HttpStatus, response)
	}

	//  size
	p.Width = uint32(img.Width())
	p.Height = uint32(img.Height())

	// URL
	if config.GetBool("http.tlsEnabled") {
		p.URL = fmt.Sprintf("https://%s/api/v1/photo/%s/max", config.GetStringP("hostname"), p.Hash)
	} else {
		p.URL = fmt.Sprintf("http://%s/api/v1/photo/%s/max", config.GetStringP("hostname"), p.Hash)
	}

	// save in datastore
	if err = datastore.Put(p.Hash, photoBytes); err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - put photo in store failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "datastoreFailed"
		return c.JSON(response.HttpStatus, response)
	}

	// create entry in DB
	if err = p.Create(); err != nil {
		response.Code = "duplicate"
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - duplicate photo %s", c.RealIP(), response.UUID, p.Hash)
		response.HttpStatus = http.StatusConflict

		// remove photo from datastore
		if !strings.HasPrefix(err.Error(), "UNIQUE") {
			if err = datastore.Delete(p.Hash); err != nil {
				c.LogErrorf(" datastore.Delete(%s): %v", p.Hash, err)
			}
			response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - photo.Create failed: %v", c.RealIP(), response.UUID, err)
			response.HttpStatus = http.StatusInternalServerError
			response.Code = "dbCreateFailed"

		}
		log.Error(response.Message)
		return c.JSON(response.HttpStatus, response)
	}

	// marshal photo
	response.Data, err = json.Marshal(p)
	if err != nil {
		response.Message = fmt.Sprintf("%v - %s - handlers.PhotoCreate - json.Marshal(photo) failed: %v", c.RealIP(), response.UUID, err)
		log.Error(response.Message)
		response.HttpStatus = http.StatusInternalServerError
		response.Code = "marshalFailed"
		return c.JSON(response.HttpStatus, response)
	}

	// GG
	response.Success = true
	response.HttpStatus = http.StatusCreated
	return c.JSON(response.HttpStatus, response)
}

// PhotoGetProperties returns PhotoProperties
func PhotoGetProperties(c echo.Context) error {
	// get ID -> hash
	hash := c.Param("id")
	// get photo
	phot, err := photo.GetByHash(hash)
	if err != nil {
		log.Infof("DEBUG err %v", err)
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("%v - controllers.PhotoGetProperties - unable to photo.GetByHash(%s): %v", c.RealIP(), hash, err)
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
	case sql.ErrNoRows:
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
	response.Photo = photoOri
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
