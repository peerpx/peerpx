package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/core"
)

// Category temp definition
type Category uint8

// Tag temp
type Tag string

// Licence temp
type Licence uint8

// Comment temp
type Comment struct {
}

/*
// PhotoPublicProperties respo
type PhotoProperties struct {
	Hash         string          `json:"hash"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	camera       string          `json:"camera"`
	Lens         string          `json:"lens"`
	FocalLength  uint16          `json:"focalLength"`
	Iso          uint16          `json:"iso"`
	ShutterSpeed string          `json:"shutterSpeed"`
	Aperture     float32         `json:"aperture"`
	TimeViewed   uint64          `json:"timeViewed"`
	Rating       float32         `json:"rating"`
	Category     models.Category `json:"category"`
	Location     string          `json:"location"`
	Privacy      bool            `json:"privacy"` // true if private
	Latitude     float32         `json:"latitude"`
	Longitude    float32         `json:"longitude"`
	TakenAt      time.Time       `json:"takenAt"`
	Width        uint32          `json:"width"`
	Height       uint32          `json:"height"`
	Nsfw         bool            `json:"nsfw"`
	LicenceType  models.Licence  `json:"licenceType"`
	URL          string          `json:"url"`
	User         string          `json:"user"` // @user@instance
	Tags         []models.Tag    `json:"tags"`
}
*/

// Photo represents a Photo
type Photo struct {
	gorm.Model   `json:"-"`
	Hash         string `gorm:"type:varchar(100);unique_index"` // sha256 + base58 ?
	Name         string
	Description  string
	Camera       string
	Lens         string
	FocalLength  uint16
	Iso          uint16
	ShutterSpeed string  // or float ? "1/250" vs 0.004
	Aperture     float32 // 5.6, 32, 1.4
	TimeViewed   uint64
	Rating       float32
	Category     Category
	Location     string
	Privacy      bool // true if private
	Latitude     float32
	Longitude    float32
	TakenAt      time.Time
	Width        uint32
	Height       uint32
	Nsfw         bool
	LicenceType  Licence
	URL          string
	User         User
	Comments     []Comment `gorm:"-"`
	Tags         []Tag     `gorm:"-"`
}

// PhotoGetByHash return photo from its hash
func PhotoGetByHash(hash string) (photo Photo, err error) {
	err = core.DB.Find(&photo).Where("hash = ?", hash).Error
	// todo load user
	// todo load tags ?
	return
}

// PhotoDeleteByHash delete photo from DB and datastore
// we don't care if photo is not found
func PhotoDeleteByHash(hash string) error {
	err := core.DB.Unscoped().Where("hash = ?").Delete(Photo{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	err = core.DS.Delete(hash)
	if err != nil && err != core.ErrNotFoundInDatastore {
		return err
	}
	return nil
}

// PhotoList list photos regarding optionnal args
func PhotoList(args ...interface{}) (photos []Photo, err error) {
	err = core.DB.Find(&photos).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return photos, err
	}
	return
}

// Create save new photo in DB
func (p *Photo) Create() error {
	return core.DB.Create(p).Error
}

// Update update photo in DB
func (p *Photo) Update() error {
	if p.ID == 0 {
		return errors.New("not DBifiée")
	}
	return core.DB.Update(p).Error
}

//func (p *Photo) GetPublicProperties() {}
