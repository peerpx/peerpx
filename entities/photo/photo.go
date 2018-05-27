package photo

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/datastore"
	"github.com/peerpx/peerpx/services/db"
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
	User         user.User
	// Todo remove
	Comments []Comment `gorm:"-" json:"-"`
	Tags     []Tag     `gorm:"-"`
}

// GetByHash return photo from its hash
func GetByHash(hash string) (photo Photo, err error) {
	err = db.DB.Find(&photo).Where("hash = ?", hash).Error
	// todo load user
	// todo load tags ?
	return
}

// DeleteByHash delete photo from DB and datastore
// we don't care if photo is not found
func DeleteByHash(hash string) error {
	err := db.DB.Unscoped().Where("hash = ?", hash).Delete(Photo{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	err = datastore.Delete(hash)
	if err != nil && err != datastore.ErrNotFound {
		return err
	}
	return nil
}

// List list photos regarding optional args
func List(args ...interface{}) (photos []Photo, err error) {
	err = db.DB.Find(&photos).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return photos, err
	}
	return
}

// Create save new photo in DB
func (p *Photo) Create() error {
	return db.DB.Create(p).Error
}

// Update update photo in DB
func (p *Photo) Update() error {
	if p.ID == 0 {
		return errors.New("photo is not recoded in DB yet, i can't update it")
	}
	return db.DB.Save(p).Error
}

// Validate check if photo properties are valid
// 0: ok
// 1: Name is too long (max length: 255)
// 2: Camera is too long (max length: 255)
// 3: Lens is too long (max length: 255)
// 4: ShutterSpeed is too long (max length: 255)
// 5: Location is too long (max length: 255)
// 6: Latitude is out of range ( -90.00 < latitude < +90.00
// 7: Longitude is out of range (-180 < longitude < +180.00
// 8: Hey Marty "TakenAt" is in the future !
func (p *Photo) Validate() uint8 {
	// Name is stored as varchar(255)
	if len(p.Name) > 255 {
		return 1
	}

	// Camera varChar(255)
	if len(p.Camera) > 255 {
		return 2
	}

	// Lens  varChar(255)
	if len(p.Lens) > 255 {
		return 3
	}

	// ShutterSpeed string  // or float ? "1/250" vs 0.004
	if len(p.ShutterSpeed) > 255 {
		return 4
	}

	// TODO	Category Category

	//	Location varchar(255)
	if len(p.Location) > 255 {
		return 5
	}

	//Latitude
	if p.Latitude < -90.00 || p.Latitude > 90.00 {
		return 6
	}

	// Longitude
	if p.Longitude < -180.00 || p.Longitude > 180.00 {
		return 7
	}

	// TakenAt      time.Time not in future
	if p.TakenAt.After(time.Now()) {
		return 8
	}

	// TODO LicenceType  Licence

	// OK
	return 0
}
