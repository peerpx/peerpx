package user

import (
	"errors"
	"net/mail"

	"fmt"

	"strings"

	"unicode/utf8"

	"github.com/jinzhu/gorm"
	"github.com/peerpx/peerpx/services/db"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// User represent an user
type User struct {
	gorm.Model `json:"-"`
	Username   string `gorm:"type:varchar(255);unique_index"`
	Email      string `gorm:"unique_index"`
	Password   string
	Firstname  string
	Lastname   string
	Gender     Gender
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

// Create creates and returns a new user
func Create(email, username, clearPassword string) (user *User, err error) {
	// validate entries
	// email
	email = strings.ToLower(email)
	if _, err = mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("%s is not a valid email", email)
	}
	// username length
	username = strings.ToLower(username)
	usernameLength := utf8.RuneCountInString(username)
	if usernameLength > viper.GetInt("usernameMaxLength") {
		return nil, fmt.Errorf("username must have %d char max", viper.GetInt("usernameMaxLength"))
	}
	if usernameLength < viper.GetInt("usernameMinLength") {
		return nil, fmt.Errorf("username must have %d char min", viper.GetInt("usernameMinLength"))
	}

	// password
	if utf8.RuneCountInString(clearPassword) < viper.GetInt("passwordMinLength") {
		return nil, fmt.Errorf("password must be at least %d char long", viper.GetInt("passwordMinLength"))
	}

	user = new(User)
	user.Username = username
	user.Email = email
	passwordByte, err := bcrypt.GenerateFromPassword([]byte(clearPassword), 10)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %v", err)
	}
	user.Password = string(passwordByte)

	// create
	if err = user.Create(); err != nil {
		return nil, fmt.Errorf("unable to record new user in database: %v", err)
	}
	return user, nil
}

// UserGetByID return user by its ID
func UserGetByID(id int) (user *User, err error) {
	user = new(User)
	err = db.DB.Find(user).Where("id = ?", id).Error
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
