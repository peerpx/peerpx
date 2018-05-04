package models

import (
	"time"

	"errors"
	"github.com/jinzhu/gorm"
	"github.com/toorop/peerpx/core"
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

// Photo represents a Photo
type Photo struct {
	gorm.Model
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

// Resize resize photo
func (p *Photo) Resize(w, h uint) error {
	return nil
}

// ResizeByHeight resize photo by height
func (p *Photo) ResizeByHeight(h uint) error {
	return p.Resize(0, h)
}

// ResizeByWidth resize photo by width
func (p *Photo) ResizeByWidth(w uint) error {
	return p.Resize(w, 0)
}
