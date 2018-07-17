package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	// import image
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/context"
	"github.com/peerpx/peerpx/entities/photo"
	"github.com/peerpx/peerpx/entities/user"
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
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	// get multipart
	form, err := c.MultipartForm()
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - c.MultipartForm() failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusBadRequest, "reqNotMultipart", msg)
	}

	// get photo properties
	photoProperties := form.Value["properties"]
	if len(photoProperties) != 1 {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - len(form.Value[properties]) != 1", c.RealIP(), response.UUID)
		return response.Error(c, http.StatusBadRequest, "reqBadMultipartProperties", msg)

	}

	// TODO verifier les props

	// unmarshall photo
	p := new(photo.Photo)
	if err := json.Unmarshal([]byte(photoProperties[0]), p); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - umarshall form.Value[properties] failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusBadRequest, "reqBadPhotoProperties", msg)

	}

	// get raw photo ("file")
	photoFile := form.File["file"]
	if len(photoFile) != 1 {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - len(form.File[file]) != 1", c.RealIP(), response.UUID)
		return response.Error(c, http.StatusBadRequest, "reqBadMultipartFile", msg)
	}

	// open photo file
	fd, err := photoFile[0].Open()
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - photoFile[0].Open() failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "openFormFileFailed", msg)
	}
	defer fd.Close()

	// read photo file
	photoBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - ioutil.ReadAll(photoFile) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "readFormFileFailed", msg)
	}

	// check mime type
	mimeType := http.DetectContentType(photoBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - %s is not supported for photo", c.RealIP(), response.UUID, mimeType)
		return response.Error(c, http.StatusBadRequest, "unsupportedPhotoFormat", msg)
	}

	// resize && re-encode
	img, err := image.NewFromBytes(photoBytes)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - image.NewFromBytes(photoBytes) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "imageNewFailed", msg)
	}

	if img.Width() > config.GetIntDefault("photo.maxWidth", 2000) || img.Height() > config.GetIntDefault("photo.maxHeight", 2000) {
		err = img.ResizeToFit(config.GetIntDefault("photo.maxWidth", 2000), config.GetIntDefault("photo.maxHeight", 2000))
		if err != nil && err != image.ErrUpscaleNotAllowed {
			msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - img.ResizeToFit failed: %v", c.RealIP(), response.UUID, err)
			return response.Error(c, http.StatusInternalServerError, "resizeFailed", msg)
		}
	}

	// To JPEG
	photoBytes, err = img.JPEG(100)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - img.JPEG failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "conversionJpegFailed", msg)
	}

	// get hash
	p.Hash, err = hasher.GetHash(photoBytes)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - hasher.GetHash(photoBytes) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "conversionJpegFailed", msg)
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

	// get user
	ui := c.Get("u")
	if ui == nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - c.Get(u) return empty string.", c.RealIP(), response.UUID)
		return response.Error(c, http.StatusUnauthorized, "userNotInContext", msg)
	}
	u := ui.(*user.User)
	p.UserID = u.ID

	// save in datastore
	if err = datastore.Put(p.Hash, photoBytes); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - put photo in store failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "datastoreFailed", msg)
	}

	// create entry in DB
	if err = p.Create(); err != nil {
		response.Code = "duplicate"
		response.Message = fmt.Sprintf("%s - %s - handlers.PhotoCreate - duplicate photo %s", c.RealIP(), response.UUID, p.Hash)
		response.HttpStatus = http.StatusConflict

		// remove photo from datastore
		if !strings.HasPrefix(err.Error(), "UNIQUE") {
			if err2 := datastore.Delete(p.Hash); err2 != nil {
				c.LogErrorf(" datastore.Delete(%s): %v", p.Hash, err2)
			}
			response.Message = fmt.Sprintf("%s - %s - handlers.PhotoCreate - photo.Create failed: %v", c.RealIP(), response.UUID, err)
			response.HttpStatus = http.StatusInternalServerError
			response.Code = "dbCreateFailed"

		}
		log.Error(response.Message)
		return c.JSON(response.HttpStatus, response)
	}

	// marshal photo
	response.Data, err = json.Marshal(p)
	if err != nil {
		response.Data = nil
		msg := fmt.Sprintf("%s - %s - handlers.PhotoCreate - json.Marshal(photo) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "marshalFailed", msg)
	}
	return response.OK(c, http.StatusCreated)
}

// PhotoGetProperties returns PhotoProperties
func PhotoGetProperties(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	// get ID -> hash
	hash := c.Param("id")

	// get photo
	p, err := photo.GetByHash(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Code = "notFound"
			response.HttpStatus = http.StatusNotFound
			return c.JSON(response.HttpStatus, response)
		}
		msg := fmt.Sprintf("%s - %s - handlers.PhotoGetProperties - photo.GetByHash(%s) failed: %v", c.RealIP(), response.UUID, hash, err)
		return response.Error(c, http.StatusInternalServerError, "getByHashFailed", msg)
	}
	// marshal photo
	response.Data, err = json.Marshal(p)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoGetProperties - json.Marshal(photo) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "marshalFailed", msg)
	}

	return response.OK(c, http.StatusOK)
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
	// cache
	c.Response().Header().Set("Etag", hash)
	c.Response().Header().Set("Cache-Control", "max-age=120")

	return c.Blob(http.StatusOK, "image/jpeg ", photoBytes)
}

// PhotoPut alter photo properties
func PhotoPut(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	// read body
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoPut - read request body failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "readBodyFailed", msg)
	}

	// -> photoNew
	var photoNew photo.Photo
	if err := json.Unmarshal(bodyBytes, &photoNew); err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoPut - unmarshal body failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "unmarshallBodyFailed", msg)
	}

	// validate
	if status := photoNew.Validate(); status != 0 {
		return response.Error(c, http.StatusBadRequest, fmt.Sprintf("errValidationFailed_%d", status), "")
	}

	// get photo props ->  photoOri
	photoOri, err := photo.GetByHash(photoNew.Hash)
	switch err {
	case sql.ErrNoRows:
		return response.Error(c, http.StatusNotFound, "errNotFound", "")
	case nil:
	default:
		msg := fmt.Sprintf("%s - %s - handlers.PhotoPut - photo.GetByHash(%s) failed: %v", c.RealIP(), response.UUID, photoNew.Hash, err)
		return response.Error(c, http.StatusInternalServerError, "photoByHashFailed", msg)
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
		msg := fmt.Sprintf("%s - %s - handlers.PhotoPut - photo.Update failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "photoUpdateFailed", msg)
	}

	// return photo
	response.Data, err = json.Marshal(photoOri)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoPut - json.Marshal(photo) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "photoMarshalFailed", msg)
	}
	return response.OK(c, http.StatusOK)
}

// PhotoDel delete a photo
func PhotoDel(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	// get hash
	hash := c.Param("id")
	if err := photo.DeleteByHash(hash); err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, http.StatusNotFound, "notFound", "")
		}
		msg := fmt.Sprintf("%s - %s - handlers.PhotoDel - photo.DeleteByHash(%s) failed: %v", c.RealIP(), response.UUID, hash, err)
		return response.Error(c, http.StatusInternalServerError, "photoDeleteByHashFailed", msg)
	}
	return response.OK(c, http.StatusOK)
}

// PhotoResize returns resized photo
func PhotoResize(c echo.Context) error {
	hash := c.Param("id")

	// cache
	if IfNoneMatch := c.Request().Header.Get("if-none-match"); IfNoneMatch == hash {
		// check if the photo still exists in datastore
		exists, _ := datastore.Exists(hash)
		if exists {
			c.Response().Header().Set("Etag", hash)
			c.Response().Header().Set("Cache-Control", "max-age=3600")
			return c.NoContent(http.StatusNotModified)
		}
	}

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

	imgBytes, err := datastore.Get(hash)
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

	// cache
	c.Response().Header().Set("Etag", hash)
	c.Response().Header().Set("Cache-Control", "max-age=3600")

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
func PhotoSearch(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewApiResponse(c.UUID)

	photos, err := photo.List()
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoSearch - photo.List() failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "photoListFailed", msg)
	}
	data := PhotoSearchResponse{
		Total:  len(photos),
		Limit:  0,
		Offset: 0,
		Data:   photos,
	}

	response.Data, err = json.Marshal(data)
	if err != nil {
		msg := fmt.Sprintf("%s - %s - handlers.PhotoSearch - json.Marshal(data) failed: %v", c.RealIP(), response.UUID, err)
		return response.Error(c, http.StatusInternalServerError, "marshalFailed", msg)
	}

	return response.OK(c, http.StatusOK)
}
