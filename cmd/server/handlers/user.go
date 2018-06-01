package handlers

import (
	"net/http"

	"io/ioutil"

	"encoding/json"

	"github.com/labstack/echo"
	"github.com/peerpx/peerpx/cmd/server/middlewares"
	"github.com/peerpx/peerpx/entities/user"
	"github.com/peerpx/peerpx/services/log"
)

type userCreateRequest struct {
	Email    string
	Username string
	Password string
}

type userCreateResponse struct {
	User *user.User `json:",omitempty"`
	Msg  string     `json:",omitempty"`
}

// UserCreate create a new user
func UserCreate(c echo.Context) error {
	response := new(userCreateResponse)
	// get body
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Errorf("%v - handlers.UserAdd - unable to read request body: %v", c.RealIP(), err)
		response.Msg = "bad request body"
		return c.JSON(http.StatusBadRequest, response)
	}
	// unmarshal
	requestData := new(userCreateRequest)
	if err = json.Unmarshal(body, requestData); err != nil {
		log.Errorf("%v - handlers.UserAdd - unable to unmarshall request body: %v", c.RealIP(), err)
		response.Msg = "bad json"
		return c.JSON(http.StatusBadRequest, response)
	}

	response.User, err = user.Create(requestData.Email, requestData.Username, requestData.Password)
	if err != nil {
		log.Errorf("%v - handlers.UserAdd - unable to create user: %v", c.RealIP(), err)
		response.Msg = err.Error()
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusCreated, response)
}

type userLoginRequest struct {
	Login    string
	Password string
}

type userLoginResponse struct {
	User *user.User `json:",omitempty"`
	Msg  string     `json:",omitempty"`
}

// UserLogin used to login
func UserLogin(ac echo.Context) error {
	c := ac.(*middlewares.AppContext)
	response := new(userLoginResponse)
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Errorf("%v - handlers.UserLogin - unable to read request body: %v", c.RealIP(), err)
		response.Msg = "bad request body"
		return c.JSON(http.StatusBadRequest, response)
	}
	// unmarshall
	requestData := new(userLoginRequest)
	if err = json.Unmarshal(body, requestData); err != nil {
		log.Errorf("%v - handlers.UserLogin - unable to unmarshall request body: %v", c.RealIP(), err)
		response.Msg = "bad json"
		return c.JSON(http.StatusBadRequest, response)
	}

	u, err := user.Login(requestData.Login, requestData.Password)
	if err != nil {
		if err == user.ErrNoSuchUser {
			log.Errorf("%v - handlers.UserLogin - unable to login: %v", c.RealIP(), err)
			response.Msg = "no such user"
			return c.JSON(http.StatusUnauthorized, response)
		}
		response.Msg = "unable to login"
		return c.JSON(http.StatusInternalServerError, response)
	}
	// set user in session
	if err = c.SessionSet("username", u.Username); err != nil {
		log.Errorf("%s -  handlers.UserLogin - unable to put username %s in session: %v", c.Request().RemoteAddr, u.Username, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	response.User = u
	return c.JSON(http.StatusOK, response)
}

// a re-utiliser pour le PUT
/*
// UserPostRequest is request struct for adding user
type UserPostRequest struct {
	Username  string
	Firstname string
	Lastname  string
	Gender    user.Gender
	Email     string
	Address   string
	City      string
	State     string
	Zip       string
	Country   string
	About     string
	Locale    string // char(2)
	ShowNsfw  bool
}

// UserPostResponse is response on adding user request
type UserPostResponse struct {
	Code uint8
	Msg  string
	User user.User
}

// UserPost handle POST /api/v1/user request
func User(c echo.Context) error {
	response := UserPostResponse{}

	// get body request
	userDatas, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Infof("%v - handlers.UserPost - unable to read request body: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer c.Request().Body.Close()

	// decode json request in UserPostRequest struct
	var userpost UserPostRequest
	err = json.Unmarshal(userDatas, &userpost)
	if err != nil {
		log.Infof("%v - handlers.UserPost - unable to unmarshall json from body: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// validating mail address
	if _, err = mail.ParseAddress(userpost.Email); err != nil {
		log.Infof("%v - handlers.UserPost - invalid mail address: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// validating mail address
	if len(userpost.Username) < 1 {
		log.Infof("%v - handlers.UserPost - invalid username: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// creating user for saving in DB
	user := user.User{
		Username:  userpost.Username,
		Firstname: userpost.Firstname,
		Lastname:  userpost.Lastname,
		Gender:    userpost.Gender,
		Email:     userpost.Email,
		Address:   userpost.Address,
		City:      userpost.City,
		State:     userpost.State,
		Zip:       userpost.Zip,
		Country:   userpost.Country,
		About:     userpost.About,
		Locale:    userpost.Locale,
		ShowNsfw:  userpost.ShowNsfw,
		UserURL:   "",
		Admin:     false,
		AvatarURL: "",
		APIKey:    "",
	}

	if err := user.Create(); err != nil {
		log.Infof("%v - handlers.UserPost - unable to create user in DB: %v", c.RealIP(), err)
		return c.NoContent(http.StatusInternalServerError)
	}

	response.User = user
	return c.JSON(http.StatusCreated, response)
}
*/
