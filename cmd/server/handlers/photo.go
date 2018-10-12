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
	response := NewAPIResponse(c)

	// get multipart
	form, err := c.MultipartForm()
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - c.MultipartForm() failed: %v", err)
		response.Code = "reqNotMultipart"
		return response.KO(http.StatusBadRequest)
	}

	// get photo properties
	photoProperties := form.Value["properties"]
	if len(photoProperties) != 1 {
		response.Log = "handlers.PhotoCreate - len(form.Value[properties]) != 1"
		response.Code = "reqBadMultipartProperties"
		return response.KO(http.StatusBadRequest)

	}

	// TODO verifier les props

	// unmarshall photo
	p := new(photo.Photo)
	if err := json.Unmarshal([]byte(photoProperties[0]), p); err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - umarshall form.Value[properties] failed: %v", err)
		response.Code = "reqBadPhotoProperties"
		return response.KO(http.StatusBadRequest)

	}

	// get raw photo ("file")
	photoFile := form.File["file"]
	if len(photoFile) != 1 {
		response.Log = "handlers.PhotoCreate - len(form.File[file]) != 1"
		response.Code = "reqBadMultipartFile"
		return response.KO(http.StatusBadRequest)
	}

	// open photo file
	fd, err := photoFile[0].Open()
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - photoFile[0].Open() failed: %v", err)
		response.Code = "openFormFileFailed"
		return response.KO(http.StatusInternalServerError)
	}
	defer fd.Close()

	// read photo file
	photoBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - ioutil.ReadAll(photoFile) failed: %v", err)
		response.Code = "readFormFileFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// check mime type
	mimeType := http.DetectContentType(photoBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - %s is not supported for photo", mimeType)
		response.Code = "unsupportedPhotoFormat"
		return response.KO(http.StatusBadRequest)
	}

	// resize && re-encode
	img, err := image.NewFromBytes(photoBytes)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - image.NewFromBytes(photoBytes) failed: %v", err)
		response.Code = "imageNewFailed"
		return response.KO(http.StatusInternalServerError)
	}

	if img.Width() > config.GetIntDefault("photo.maxWidth", 2000) || img.Height() > config.GetIntDefault("photo.maxHeight", 2000) {
		err = img.ResizeToFit(config.GetIntDefault("photo.maxWidth", 2000), config.GetIntDefault("photo.maxHeight", 2000))
		if err != nil && err != image.ErrUpscaleNotAllowed {
			response.Log = fmt.Sprintf("handlers.PhotoCreate - img.ResizeToFit failed: %v", err)
			response.Code = "resizeFailed"
			return response.KO(http.StatusInternalServerError)
		}
	}

	// To JPEG
	photoBytes, err = img.JPEG(100)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - img.JPEG failed: %v", err)
		response.Code = "conversionJpegFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// get hash
	p.Hash, err = hasher.GetHash(photoBytes)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - hasher.GetHash(photoBytes) failed: %v", err)
		response.Code = "conversionJpegFailed"
		return response.KO(http.StatusInternalServerError)
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
		response.Log = "handlers.PhotoCreate - c.Get(u) return empty string"
		response.Code = "userNotInContext"
		return response.KO(http.StatusUnauthorized)
	}
	u := ui.(*user.User)
	p.UserID = u.ID

	// save in datastore
	if err = datastore.Put(p.Hash, photoBytes); err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - put photo in store failed: %v", err)
		response.Code = "datastoreFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// create entry in DB
	if err = p.Create(); err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoCreate - duplicate photo %s", p.Hash)
		response.Code = "duplicate"
		response.HTTPStatus = http.StatusConflict

		// remove photo from datastore
		if !strings.HasPrefix(err.Error(), "UNIQUE") {
			if err2 := datastore.Delete(p.Hash); err2 != nil {
				c.LogErrorf(" datastore.Delete(%s): %v", p.Hash, err2)
			}
			response.Log = fmt.Sprintf("handlers.PhotoCreate - photo.Create failed: %v", err)
			response.HTTPStatus = http.StatusInternalServerError
			response.Code = "dbCreateFailed"

		}
		return c.JSON(response.HTTPStatus, response)
	}

	// marshal photo
	response.Data, err = json.Marshal(p)
	if err != nil {
		response.Data = nil
		response.Log = fmt.Sprintf("handlers.PhotoCreate - json.Marshal(photo) failed: %v", err)
		response.Code = "marshalFailed"
		return response.KO(http.StatusInternalServerError)
	}
	return response.OK(http.StatusCreated)
}

// PhotoGetProperties returns PhotoProperties
func PhotoGetProperties(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	// get ID -> hash
	hash := c.Param("id")

	// get photo
	p, err := photo.GetByHash(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Code = "notFound"
			return response.KO(http.StatusNotFound)
		}
		response.Log = fmt.Sprintf("handlers.PhotoGetProperties - photo.GetByHash(%s) failed: %v", hash, err)
		response.Code = "getByHashFailed"
		return response.KO(http.StatusInternalServerError)
	}
	// marshal photo
	response.Data, err = json.Marshal(p)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoGetProperties - json.Marshal(photo) failed: %v", err)
		response.Code = "marshalFailed"
		return response.KO(http.StatusInternalServerError)
	}

	return response.OK(http.StatusOK)
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
	response := NewAPIResponse(c)

	// read body
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoPut - read request body failed: %v", err)
		response.Code = "readBodyFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// -> photoNew
	var photoNew photo.Photo
	if err := json.Unmarshal(bodyBytes, &photoNew); err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoPut - unmarshal body failed: %v", err)
		response.Code = "unmarshallBodyFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// validate
	if status := photoNew.Validate(); status != 0 {
		response.Code = fmt.Sprintf("errValidationFailed_%d", status)
		return response.KO(http.StatusBadRequest)
	}

	// get photo props ->  photoOri
	photoOri, err := photo.GetByHash(photoNew.Hash)
	switch err {
	case sql.ErrNoRows:
		response.Code = "errNotFound"
		return response.KO(http.StatusNotFound)
	case nil:
	default:
		response.Log = fmt.Sprintf("handlers.PhotoPut - photo.GetByHash(%s) failed: %v", photoNew.Hash, err)
		response.Code = "photoByHashFailed"
		return response.KO(http.StatusInternalServerError)
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
		response.Log = fmt.Sprintf("handlers.PhotoPut - photo.Update failed: %v", err)
		response.Code = "photoUpdateFailed"
		return response.KO(http.StatusInternalServerError)
	}

	// return photo
	response.Data, err = json.Marshal(photoOri)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoPut - json.Marshal(photo) failed: %v", err)
		response.Code = "photoMarshalFailed"
		return response.KO(http.StatusInternalServerError)
	}
	return response.OK(http.StatusOK)
}

// PhotoDel delete a photo
func PhotoDel(ac echo.Context) error {
	c := ac.(*context.AppContext)
	response := NewAPIResponse(c)

	// get hash
	hash := c.Param("id")
	if err := photo.DeleteByHash(hash); err != nil {
		if err == sql.ErrNoRows {
			response.Code = "notFound"
			return response.KO(http.StatusNotFound)
		}
		response.Log = fmt.Sprintf("handlers.PhotoDel - photo.DeleteByHash(%s) failed: %v", hash, err)
		response.Code = "photoDeleteByHashFailed"
		return response.KO(http.StatusInternalServerError)
	}
	return response.OK(http.StatusOK)
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
	response := NewAPIResponse(c)

	photos, err := photo.List()
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoSearch - photo.List() failed: %v", err)
		response.Code = "photoListFailed"
		return response.KO(http.StatusInternalServerError)
	}
	data := PhotoSearchResponse{
		Total:  len(photos),
		Limit:  0,
		Offset: 0,
		Data:   photos,
	}

	response.Data, err = json.Marshal(data)
	if err != nil {
		response.Log = fmt.Sprintf("handlers.PhotoSearch - json.Marshal(data) failed: %v", err)
		response.Code = "marshalFailed"
		return response.KO(http.StatusInternalServerError)
	}

	return response.OK(http.StatusOK)
}
