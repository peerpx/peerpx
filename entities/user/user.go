package user

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"database/sql"

	"github.com/peerpx/peerpx/services/config"
	"github.com/peerpx/peerpx/services/db"
	"golang.org/x/crypto/bcrypt"
)

// User represent an user
type User struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Gender    Gender `json:"gender"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
	About     string `json:"about"`
	Locale    string `json:"locale"` // char(2)
	ShowNsfw  bool   `db:"show_nsfw",json:"show_nsfw"`
	UserURL   string `db:"user_url",json:"user_url"`
	Admin     bool   `json:"admin"`
	AvatarURL string `db:"avatar_url",json:"avatar_url"`
	APIKey    string `json:"-"`
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
	if usernameLength > config.GetIntDefault("username.maxLength", 25) {
		return nil, fmt.Errorf("username must have %d char max", config.GetIntDefault("username.maxLength", 25))
	}
	if usernameLength < config.GetIntDefault("username.minLength", 4) {
		return nil, fmt.Errorf("username must have %d char min", config.GetIntDefault("username.minLength", 4))
	}

	// password
	if utf8.RuneCountInString(clearPassword) < config.GetIntDefault("password.minLength", 6) {
		return nil, fmt.Errorf("password must be at least %d char long", config.GetIntDefault("password.minLength", 6))
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

// GetByID return user by its ID
func GetByID(id int) (user *User, err error) {
	user = new(User)
	err = db.Get(user, "SELECT * FROM users WHERE id=$1", id)
	return
}

// UserGetByUsername return user by its ID
func GetByUsername(username string) (user *User, err error) {
	user = new(User)
	username = strings.TrimSpace(strings.ToLower(username))
	err = db.Get(user, "SELECT * FROM users WHERE username=$1", username)
	return
}

// GetByEmail returns user by his email
func GetByEmail(email string) (user *User, err error) {
	user = new(User)
	email = strings.TrimSpace(strings.ToLower(email))
	err = db.Get(user, "SELECT * FROM users WHERE email=?", email)
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
		switch err.Error() {
		case sql.ErrNoRows.Error():
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
	stmt, err := db.Preparex("INSERT INTO users (username,firstname,lastname,gender,email,address,city,state,zip,country,about,locale,show_nsfw,user_url,admin,avatar_url,api_key,password) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(u.Username, u.Firstname, u.Lastname, u.Gender, u.Email, u.Address, u.City, u.State, u.Zip, u.Country, u.About, u.Locale, u.ShowNsfw, u.UserURL, u.Admin, u.AvatarURL, u.APIKey, u.Password)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	u.ID = uint(id)
	return err
}

// Update update user in DB
func (u *User) Update() error {
	if u.ID == 0 {
		return errors.New("user unknown in database")
	}
	stmt, err := db.Preparex("UPDATE users SET username = ?, firstname = ?, lastname = ?, gender = ?, email = ?, address = ?, city = ?, state  = ?, zip = ?, country = ?, about = ?, locale = ?, show_nsfw = ?, user_url = ?, admin = ?, avatar_url = ?, api_key = ?, password = ? WHERE id = ?")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(u.Username, u.Firstname, u.Lastname, u.Gender, u.Email, u.Address, u.City, u.State, u.Zip, u.Country, u.About, u.Locale, u.ShowNsfw, u.UserURL, u.Admin, u.AvatarURL, u.APIKey, u.Password, u.ID)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	u.ID = uint(id)
	return err

}
