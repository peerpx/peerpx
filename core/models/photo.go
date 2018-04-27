package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Photo represents a Photo
type Photo struct {
	gorm.Model
	IPFSHash     string
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
	Images       []Image
	User         User
	Comments     []Comment
	Tags         []Tag
}

// Category temp definition
type Category uint8

// Tag temp
type Tag string

// Licence temp
type Licence uint8

// Image represents
type Image struct {
	Size uint32
	URL  string
}

// User temp
type User struct {
}

// Comment temp
type Comment struct {
}
