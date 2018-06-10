package photo

import (
	"errors"
	"time"

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

// Photo represents a Photo
type Photo struct {
	ID           uint      `json:"id"`
	Hash         string    `json:"hash"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Camera       string    `json:"camera"`
	Lens         string    `json:"lens"`
	FocalLength  uint16    `json:"focal_length"`
	Iso          uint16    `json:"iso"`
	ShutterSpeed string    `json:"shutter_speed"` // or float ? "1/250" vs 0.004
	Aperture     float32   `json:"aperture"`      // 5.6, 32, 1.4
	TimeViewed   uint64    `json:"time_viewed"`
	Rating       float32   `json:"rating"`
	Category     Category  `json:"category"`
	Location     string    `json:"location"`
	Privacy      bool      `json:"privacy"` // true if private
	Latitude     float32   `json:"latitude"`
	Longitude    float32   `json:"longitude"`
	AddedAt      time.Time `json:"added_at"`
	TakenAt      time.Time `json:"taken_at"`
	Width        uint32    `json:"width"`
	Height       uint32    `json:"height"`
	Nsfw         bool      `json:"nsfw"`
	LicenceType  Licence   `json:"licence_type"`
	URL          string    `json:"url"`
	User         user.User `json:"user"`
	// Todo remove
	Comments []Comment `json:"-"`
	Tags     []Tag     `json:"-"`
}

// GetByHash return photo from its hash
func GetByHash(hash string) (photo *Photo, err error) {
	photo = new(Photo)
	err = db.Get(photo, "SELECT * FROM photo WHERE hash = ?", hash)
	// todo load user
	// todo load tags ?
	return
}

// DeleteByHash delete photo from DB and datastore
// we don't care if photo is not found
func DeleteByHash(hash string) error {
	stmt, err := db.Preparex("DELETE FROM photo WHERE hash = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(hash)
	if err != nil {
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
	err = db.Select(photos, "SELECT * FROM photo ORDER BY id DESC")
	return
}

// Create save new photo in DB
func (p *Photo) Create() error {
	stmt, err := db.Preparex("INSERT INTO photos (added_at, hash, name, description, camera,lens,focal_length,iso, shutter_speed, aperture, time_viewed, rating, category , location, privacy, latitude, longitude, taken_at, width, height, nsfw, licence_type, url) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(time.Now(), p.Hash, p.Name, p.Description, p.Camera, p.Lens, p.FocalLength, p.Iso, p.ShutterSpeed, p.Aperture, p.TimeViewed, p.Rating, p.Category, p.Location, p.Privacy, p.Latitude, p.Longitude, p.TakenAt, p.Width, p.Height, p.Nsfw, p.LicenceType, p.URL)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = uint(id)
	return nil
}

// Update update photo in DB
func (p *Photo) Update() error {
	if p.ID == 0 {
		return errors.New("photo is not recoded in DB yet, i can't update it !")
	}
	stmt, err := db.Preparex("UPDATE photos SET added_at=?, hash=?, name=?, description=?, camera=?, lens=?, focal_length=?, iso=?, shutter_speed=?, aperture=?, time_viewed=?, rating=?, category=?, location=?, privacy=?, latitude=?, longitude=?, taken_at=?, width=?, height=?, nsfw=?, licence_type=?, url=? WHERE id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(time.Now(), p.Hash, p.Name, p.Description, p.Camera, p.Lens, p.FocalLength, p.Iso, p.ShutterSpeed, p.Aperture, p.TimeViewed, p.Rating, p.Category, p.Location, p.Privacy, p.Latitude, p.Longitude, p.TakenAt, p.Width, p.Height, p.Nsfw, p.LicenceType, p.URL, p.ID)
	return err
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
