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
	Username   string `gorm:"type:varchar(255);unique_index" json:"username"`
	Email      string `gorm:"unique_index" json:"email"`
	Password   string `json:"-"`
	Firstname  string `json:"firstname"`
	Lastname   string `json:"lastname"`
	Gender     Gender `json:"gender"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	Country    string `json:"country"`
	About      string `json:"about"`
	Locale     string `json:"locale"` // char(2)
	ShowNsfw   bool   `json:"show_nsfw"`
	UserURL    string `json:"user_url"`
	Admin      bool   `json:"admin"`
	AvatarURL  string `json:"avatar_url"`
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

// errors
var (
	ErrNoSuchUser = errors.New("no such user")
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
func GetByUsername(username string) (user *User, err error) {
	user = new(User)
	username = strings.TrimSpace(strings.ToLower(username))
	err = db.DB.Find(user).Where("username = ?", username).Error
	return
}

// GetByEmail returns user by his email
func GetByEmail(email string) (user *User, err error) {
	user = new(User)
	email = strings.TrimSpace(strings.ToLower(email))
	err = db.DB.Find(user).Where("email = ?", email).Error
	return
}

// UserLogin returns user if exists
func Login(login, password string) (user *User, err error) {
	isEmail := false
	login = strings.ToLower(login)
	_, err = mail.ParseAddress(login)
	if err != nil {
		isEmail = true
	}
	if isEmail {
		user, err = GetByEmail(login)
	} else {
		user, err = GetByUsername(login)
	}
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, ErrNoSuchUser
		default:
			return nil, err
		}
	}
	// check password
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			err = ErrNoSuchUser
		}
		return nil, err
	}
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
