package models

import "github.com/jinzhu/gorm"

// User represent an user
type User struct {
	gorm.Model
	Username  string
	Firstname string
	Lastname  string
	Sex       Sex
	Email     string
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
}

// Sex is the user sex
type Sex uint8

const (
	Undefined Sex = iota
	Male
	Female
)
