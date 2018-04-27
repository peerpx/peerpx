package models

import "github.com/jinzhu/gorm"

// User represent an user
type User struct {
	gorm.Model
	Username  string `gorm:"type:varchar(255);unique_index"`
	Firstname string
	Lastname  string
	Gender    Gender
	Email     string `gorm:"unique_index"`
	Address   string
	City      string
	State     string
	Zip       string
	Country   string
	About     string
	Locale    string // char(2)
	ShowNsfw  bool
	UserURL   string
	Admin     bool
	Avatars   []Image
	APIKey    string // temp
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
