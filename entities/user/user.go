package user

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/services/db"
)

// User represent an user
type User struct {
	gorm.Model `json:"-"`
	Username   string `gorm:"type:varchar(255);unique_index"`
	Firstname  string
	Lastname   string
	Gender     Gender
	Email      string `gorm:"unique_index"`
	Address    string
	City       string
	State      string
	Zip        string
	Country    string
	About      string
	Locale     string // char(2)
	ShowNsfw   bool
	UserURL    string
	Admin      bool
	AvatarURL  string
	APIKey     string `json:"-"`
}

// Gender is the user gender
type Gender uint8

const (
	// Undefined ?
	Undefined Gender = iota
	// Male male
	Male
	// Female female
	Female
)

// UserGetByID return user by its ID
func UserGetByID(id int) (user User, err error) {
	err = db.DB.Find(&user).Where("id = ?", id).Error
	return
}

// UserGetByUsername return user by its ID
func UserGetByUsername(username string) (user User, err error) {
	err = db.DB.Find(&user).Where("username = ?", username).Error
	return
}

// Create save new user in DB
func (u *User) Create() error {
	return db.DB.Create(u).Error
}

// Update update user in DB
func (u *User) Update() error {
	if u.ID == 0 {
		return errors.New("user unknown in database")
	}
	return db.DB.Update(u).Error
}
