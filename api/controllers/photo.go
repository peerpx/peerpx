package controllers

import (
	"fmt"
	// jpeg
	_ "image/jpeg"
	// png
	_ "image/png"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"

	"strconv"
	"time"

	"github.com/jinzhu/gorm"
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
	// - descripion TODO

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

	photo := models.Photo{}
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

// PhotoGetPropertiesResponse response for PhotoGetProperties controller
type PhotoGetPropertiesResponse struct {
	Hash         string
	Name         string
	Description  string
	camera       string
	Lens         string
	FocalLength  uint16
	Iso          uint16
	ShutterSpeed string  // or float ? "1/250" vs 0.004
	Aperture     float32 // 5.6, 32, 1.4
	TimeViewed   uint64
	Rating       float32
	Category     models.Category
	Location     string
	Privacy      bool // true if private
	Latitude     float32
	Longitude    float32
	TakenAt      time.Time
	Width        uint32
	Height       uint32
	Nsfw         bool
	LicenceType  models.Licence
	URL          string
	User         string // @user@instance
	Tags         []models.Tag
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
	// response
	response := PhotoGetPropertiesResponse{
		Hash:         hash,
		Name:         photo.Name,
		Description:  photo.Description,
		camera:       photo.Camera,
		Lens:         photo.Lens,
		FocalLength:  photo.FocalLength,
		Iso:          photo.Iso,
		ShutterSpeed: photo.ShutterSpeed,
		Aperture:     photo.Aperture,
		TimeViewed:   photo.TimeViewed,
		Rating:       photo.Rating,
		Category:     photo.Category,
		Location:     photo.Location,
		Privacy:      photo.Privacy,
		Latitude:     photo.Latitude,
		Longitude:    photo.Longitude,
		TakenAt:      photo.TakenAt,
		Width:        photo.Width,
		Height:       photo.Height,
		Nsfw:         photo.Nsfw,
		LicenceType:  photo.LicenceType,
		URL:          photo.URL,
		// todo: Warning fake props
		User: "@johndoe@peerpx.com",
		Tags: []models.Tag{"fake", "sunrise"},
	}

	return c.JSON(http.StatusOK, response)
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

// PhotoSearchResponse response structure for PhotoSearch
type PhotoSearchResponse struct {
	Total int
	Limit int
	Offset int
	Data []PhotoGetPropertiesResponse
}

// PhotoSearch return an array of photos regarding the optionnals search params (TMP)
func PhotoSearch(c echo.Context) error {
	//TODO: take account of optionnal params
	photos, err := models.PhotoList()
	if err != nil {
		log.Errorf("%v - controllers.PhotoSearch - unable to list photos: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	var properties = make([]PhotoGetPropertiesResponse, 0)
	for _, p := range photos {
		property := PhotoGetPropertiesResponse{
			Hash:         p.Hash,
			Name:         p.Name,
			Description:  p.Description,
			camera:       p.Camera,
			Lens:         p.Lens,
			FocalLength:  p.FocalLength,
			Iso:          p.Iso,
			ShutterSpeed: p.ShutterSpeed,
			Aperture:     p.Aperture,
			TimeViewed:   p.TimeViewed,
			Rating:       p.Rating,
			Category:     p.Category,
			Location:     p.Location,
			Privacy:      p.Privacy,
			Latitude:     p.Latitude,
			Longitude:    p.Longitude,
			TakenAt:      p.TakenAt,
			Width:        p.Width,
			Height:       p.Height,
			Nsfw:         p.Nsfw,
			LicenceType:  p.LicenceType,
			URL:          p.URL,
			// todo: Warning fake props
			User: "@johndoe@peerpx.com",
			Tags: []models.Tag{"fake", "sunrise"},
		}
		properties = append(properties, property)
	}
	response := PhotoSearchResponse{
		Total: len(photos),
		Limit: 0,
		Offset: 0,
		Data: properties,
	}
	return c.JSON(http.StatusOK, response)
}
